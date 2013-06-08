package chunk

import (
	"bytes"
	"compress/zlib"
	"github.com/NetherrackDev/netherrack/entity"
	"github.com/NetherrackDev/netherrack/log"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/blocks"
	"runtime"
	"time"
)

func chunkController(chunk *Chunk) {
	defer func() {
		chunk.World.getRegion(chunk.X>>5, chunk.Z>>5).removeChunk()
	}()
	chunk.generate()
	tOut := time.NewTimer(1 * time.Second)
	defer tOut.Stop()
	tick := time.NewTicker(time.Second / 10)
	defer tick.Stop()

	for {
		reset := true
		select {
		case cr := <-chunk.requests:
			//Finish lighting the chunk
			needsSave := !chunk.pendingLightOperations.IsEmpty()
			for !chunk.pendingLightOperations.IsEmpty() {
				op := chunk.pendingLightOperations.Pop().(lightOperation)
				op.Execute(chunk)
			}
			for !chunk.brokenLights.IsEmpty() {
				chunk.pendingLightOperations.Push(chunk.brokenLights.Remove())
			}
			if needsSave && chunk.pendingLightOperations.IsEmpty() {
				chunk.needsSave = true
			}
			//Grab the compressed bytes
			out := chunk.toCompressedBytes(true)
			select {
			case cr.Ret <- out:
			case <-cr.Stop:
				cr.Stop <- struct{}{}
			}
		reqs:
			for {
				select {
				case cr = <-chunk.requests:
					select {
					case cr.Ret <- out:
					case <-cr.Stop:
						cr.Stop <- struct{}{}
					}
				default:
					break reqs
				}
			}
		case req := <-chunk.lightChannel:
			if req.blockLight {
				req.ret <- chunk.BlockLight(int(req.x), int(req.y), int(req.z))
			} else {
				req.ret <- chunk.SkyLight(int(req.x), int(req.y), int(req.z))
			}
		case cwr := <-chunk.watcherJoin:
			chunk.Players[cwr.P.Name()] = cwr.P
			for _, e := range chunk.Entitys {
				if e.ID() != cwr.P.ID() {
					s := e.(entity.Spawnable)
					cwr.P.RunSync(s.CreateSpawn())
				}
			}
		case cwr := <-chunk.watcherLeave:
			//Empty join first
		empty:
			for {
				select {
				case wr := <-chunk.watcherJoin:
					chunk.Players[cwr.P.Name()] = wr.P
					for _, e := range chunk.Entitys {
						if e.ID() != wr.P.ID() {
							s := e.(entity.Spawnable)
							wr.P.RunSync(s.CreateSpawn())
						}
					}
				default:
					break empty
				}
			}
			delete(chunk.Players, cwr.P.Name())
			for _, e := range chunk.Entitys {
				if e.ID() != cwr.P.ID() {
					s := e.(entity.Spawnable)
					cwr.P.RunSync(s.CreateDespawn())
				}
			}
		case cer := <-chunk.entityJoin:
			chunk.Entitys[cer.E.ID()] = cer.E
		case cer := <-chunk.entityLeave:
			delete(chunk.Entitys, cer.E.ID())
		case msg := <-chunk.messageChannel:
			reset = false
			for _, p := range chunk.Players {
				if p.ID() != msg.ID {
					p.RunSync(msg.Msg)
				}
			}
		case f := <-chunk.eventChannel:
			reset = false
			f(chunk)
		case <-tOut.C:
			if len(chunk.Players) == 0 {
				//Try and save
				if chunk.needsSave {
					chunk.Save()
				}
				//Did someone join during save?
				posChan := make(chan *ChunkPosition)
				chunk.World.chunkKillChannel <- posChan
				if len(chunk.watcherJoin) == 0 &&
					len(chunk.entityJoin) == 0 &&
					len(chunk.entityLeave) == 0 &&
					len(chunk.requests) == 0 &&
					len(chunk.watcherLeave) == 0 {
					for _, e := range chunk.Entitys {
						pausable, ok := e.(interface {
							Pause()
						})
						if ok {
							pausable.Pause()
						}
					}
					posChan <- &ChunkPosition{chunk.X, chunk.Z}
					runtime.Goexit()
				} else {
					posChan <- nil
				}
			}
		//TODO: Disable this case if nothing is happening in this chunk
		case <-tick.C:
			reset = false
			if len(chunk.blockQueue) != 0 {
				reset = true
				blockData := make([]uint32, len(chunk.blockQueue))
				for i, bc := range chunk.blockQueue {
					blockData[i] = (uint32(bc.Meta) & 0xf) | (uint32(bc.Block) << 4) | (uint32(bc.Y) << 16) | (uint32(bc.Z) << 24) | (uint32(bc.X) << 28)
				}
				for _, p := range chunk.Players {
					p.RunSync(func(s soulsand.SyncEntity) {
						sPlayer := s.(soulsand.SyncPlayer)
						sPlayer.Connection().WriteMultiBlockChange(chunk.X, chunk.Z, blockData)
					})
				}
				chunk.blockQueue = chunk.blockQueue[0:0]
			}
			needsSave := !chunk.pendingLightOperations.IsEmpty()
			count := 0
			for !chunk.pendingLightOperations.IsEmpty() {
				op := chunk.pendingLightOperations.Pop().(lightOperation)
				op.Execute(chunk)
				count++
				if count == 500 {
					runtime.Gosched()
					count = 0
				}
			}
			for !chunk.brokenLights.IsEmpty() {
				chunk.pendingLightOperations.Push(chunk.brokenLights.Remove())
				count++
				if count == 500 {
					runtime.Gosched()
					count = 0
				}
			}
			if needsSave && chunk.pendingLightOperations.IsEmpty() {
				chunk.needsSave = true
			}
		}
		if reset {
			if !tOut.Reset(1 * time.Second) {
				tOut = time.NewTimer(1 * time.Second)
			}
		}
	}
}

var lightBlockMap = []byte{
	blocks.Air.Id(),                 //0
	blocks.RedMushroom.Id(),         //1
	blocks.BrownMushroom.Id(),       //2
	blocks.Lever.Id(),               //3
	blocks.SignPost.Id(),            //4
	blocks.SignWall.Id(),            //5
	blocks.Carrots.Id(),             //6
	blocks.DeadBush.Id(),            //7
	blocks.NetherWart.Id(),          //8
	blocks.IronPressurePlate.Id(),   //9
	blocks.Rose.Id(),                //10
	blocks.Dandelion.Id(),           //11
	blocks.TallGrass.Id(),           //12
	blocks.Rails.Id(),               //13
	blocks.WoodenPressurePlate.Id(), //14
	blocks.SugarCane.Id(),           //15
}

func (chunk *Chunk) toCompressedBytes(full bool) [][]byte {
	var b bytes.Buffer
	b.Grow(5000) //Average size of compressed chunk
	w, err := zlib.NewWriterLevel(&b, zlib.BestSpeed)
	if err != nil {
		log.Println(err)
		return nil
	}
	var mask uint16 = 0
	//Blocks
	for i := 0; i < 16; i++ {
		if chunk.SubChunks[i] != nil {
			w.Write(chunk.SubChunks[i].Type)
			mask |= 1 << uint(i)
		} else if full {
			w.Write(emptySection.Type)
			mask |= 1 << uint(i)
		}
	}
	//Metadata
	for i := 0; i < 16; i++ {
		if chunk.SubChunks[i] != nil {
			w.Write(chunk.SubChunks[i].MetaData)
		} else if full {
			w.Write(emptySection.MetaData)
		}
	}
	//BlockLight
	for i := 0; i < 16; i++ {
		if chunk.SubChunks[i] != nil {
			w.Write(chunk.SubChunks[i].BlockLight)
		} else if full {
			w.Write(emptySection.BlockLight)
		}
	}
	//Skylight
	for i := 0; i < 16; i++ {
		if chunk.SubChunks[i] != nil {
			w.Write(chunk.SubChunks[i].SkyLight)
		} else if full {
			w.Write(emptySection.SkyLight)
		}
	}
	//Biomes
	w.Write(chunk.biome)
	w.Close()
	data := b.Bytes()
	out := make([]byte, 1+4+4+1+2+2+4)
	out[0] = 0x33
	writeInt(out[1:5], chunk.X)
	writeInt(out[5:9], chunk.Z)
	writeBool(out[9:10], true)
	writeUShort(out[10:12], mask)
	writeUShort(out[12:14], 0)
	writeInt(out[14:18], int32(len(data)))

	return [][]byte{out, data}
}
