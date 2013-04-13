package chunk

import (
	"bitbucket.org/Thinkofdeath/netherrack/entity"
	"bitbucket.org/Thinkofdeath/soulsand"
	"bytes"
	"compress/zlib"
	"log"
	"runtime"
	"time"
)

func chunkControler(chunk *Chunk) {
	defer func() {
		chunk.World.chunkKillChannel <- &ChunkPosition{chunk.X, chunk.Z}
	}()
	generateChunk(chunk)
	tOut := time.NewTimer(30 * time.Second)
	defer tOut.Stop()
	tick := time.NewTicker(time.Second / 20)
	defer tick.Stop()
	for {
		select {
		case cr := <-chunk.requests:
			out := chunkToCompressedBytes(chunk)
			cr.Ret <- out
		reqs:
			for {
				select {
				case cr = <-chunk.requests:
					cr.Ret <- out
				default:
					break reqs
				}
			}
		case cwr := <-chunk.watcherJoin:
			chunk.Players[cwr.P.GetID()] = cwr.P
			for _, e := range chunk.Entitys {
				if e.GetID() != cwr.P.GetID() {
					s := e.(entity.Spawnable)
					cwr.P.RunSync(s.CreateSpawn())
				}
			}
		case cwr := <-chunk.watcherLeave:
			delete(chunk.Players, cwr.P.GetID())
			for _, e := range chunk.Entitys {
				if e.GetID() != cwr.P.GetID() {
					s := e.(entity.Spawnable)
					cwr.P.RunSync(s.CreateDespawn())
				}
			}
		case cer := <-chunk.entityJoin:
			chunk.Entitys[cer.E.GetID()] = cer.E
		case cer := <-chunk.entityLeave:
			delete(chunk.Entitys, cer.E.GetID())
		case msg := <-chunk.messageChannel:
			for _, p := range chunk.Players {
				if p.GetID() != msg.ID {
					p.RunSync(msg.Msg)
				}
			}
		case f := <-chunk.eventChannel:
			f(chunk)
		case <-tOut.C:
			if len(chunk.Players) == 0 &&
				len(chunk.watcherJoin) == 0 &&
				len(chunk.Entitys) == 0 &&
				len(chunk.eventChannel) == 0 &&
				len(chunk.entityJoin) == 0 &&
				len(chunk.entityLeave) == 0 &&
				len(chunk.messageChannel) == 0 &&
				len(chunk.requests) == 0 &&
				len(chunk.watcherLeave) == 0 {
				runtime.Goexit()
			}
		case <-tick.C:
			if len(chunk.blockQueue) != 0 {
				blockData := make([]uint32, len(chunk.blockQueue))
				for i, bc := range chunk.blockQueue {
					blockData[i] = (uint32(bc.Meta) & 0xf) | (uint32(bc.Block) << 4) | (uint32(bc.Y) << 16) | (uint32(bc.Z) << 24) | (uint32(bc.X) << 28)
				}
				for _, p := range chunk.Players {
					p.RunSync(func(s soulsand.SyncPlayer) {
						s.GetConnection().WriteMultiBlockChange(chunk.X, chunk.Z, blockData)
					})
				}
				chunk.blockQueue = chunk.blockQueue[0:0]
			}
		}
		tOut.Reset(10 * time.Second)
	}
}

func chunkToCompressedBytes(chunk *Chunk) [][]byte {
	numSubs := 0
	for i := 0; i < 16; i++ {
		if chunk.SubChunks[i] != nil {
			numSubs++
		}
	}
	uncompData := make([]byte, (16*16*16+16*16*8*4)*numSubs+256)
	var mask uint16 = 0
	pSize := 16 * 16 * 16 * numSubs
	rI := 0
	for i := 0; i < 16; i++ {
		if chunk.SubChunks[i] != nil {
			copy(uncompData[16*16*16*rI:16*16*16*(rI+1)], chunk.SubChunks[i].Type)
			copy(uncompData[pSize+16*16*8*rI:pSize+16*16*8*(rI+1)], chunk.SubChunks[i].MetaData)
			copy(uncompData[pSize+(pSize/2)+16*16*8*rI:pSize+(pSize/2)+16*16*8*(rI+1)], chunk.SubChunks[i].BlockLight)
			copy(uncompData[pSize*2+16*16*8*rI:pSize*2+16*16*8*(rI+1)], chunk.SubChunks[i].SkyLight)
			mask |= 1 << uint(i)
			rI++
		}
	}
	copy(uncompData[(16*16*16+16*16*8*4)*rI:], chunk.Biome)
	var b bytes.Buffer
	w, err := zlib.NewWriterLevel(&b, zlib.BestSpeed)
	if err != nil {
		log.Println(err)
		return nil
	}
	w.Write(uncompData[:(16*16*16+16*16*8*4)*rI+256])
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
