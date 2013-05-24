package chunk

import (
	"bytes"
	"compress/zlib"
	"github.com/NetherrackDev/netherrack/entity"
	"github.com/NetherrackDev/soulsand"
	"log"
	"runtime"
	"time"
)

func chunkController(chunk *Chunk) {
	defer func() {
		chunk.World.getRegion(chunk.X>>5, chunk.Z>>5).removeChunk()
	}()
	chunk.generate()
	tOut := time.NewTimer(30 * time.Second)
	defer tOut.Stop()
	tick := time.NewTicker(time.Second / 10)
	defer tick.Stop()

	for {
		reset := true
		select {
		case cr := <-chunk.requests:
			if chunk.needsRelight {
				chunk.Relight()
			}
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
			for _, p := range chunk.Players {
				if p.ID() != msg.ID {
					p.RunSync(msg.Msg)
				}
			}
		case f := <-chunk.eventChannel:
			f(chunk)
		case <-tOut.C:
			if len(chunk.Players) == 0 {
				//Try and save
				if chunk.needsSave {
					chunk.Save()
				}
				//Did someone join during save
				posChan := make(chan *ChunkPosition)
				chunk.World.chunkKillChannel <- posChan
				if len(chunk.watcherJoin) == 0 &&
					len(chunk.eventChannel) == 0 &&
					len(chunk.entityJoin) == 0 &&
					len(chunk.entityLeave) == 0 &&
					len(chunk.messageChannel) == 0 &&
					len(chunk.requests) == 0 &&
					len(chunk.watcherLeave) == 0 {
					posChan <- &ChunkPosition{chunk.X, chunk.Z}
					runtime.Goexit()
				} else {
					posChan <- nil
				}
			}
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
			if chunk.needsRelight {
				chunk.Relight()
			}
		}
		if reset {
			tOut.Reset(1 * time.Second)
		}
	}
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
		}
	}
	//Metadata
	for i := 0; i < 16; i++ {
		if chunk.SubChunks[i] != nil {
			w.Write(chunk.SubChunks[i].MetaData)
		}
	}
	//BlockLight
	for i := 0; i < 16; i++ {
		if chunk.SubChunks[i] != nil {
			w.Write(chunk.SubChunks[i].BlockLight)
		}
	}
	//Skylight
	for i := 0; i < 16; i++ {
		if chunk.SubChunks[i] != nil {
			w.Write(chunk.SubChunks[i].SkyLight)
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
