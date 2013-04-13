package player

import (
	"bitbucket.org/Thinkofdeath/netherrack/entity"
	"bitbucket.org/Thinkofdeath/netherrack/event"
	"bitbucket.org/Thinkofdeath/netherrack/items"
	"bitbucket.org/Thinkofdeath/netherrack/nbt"
	"bitbucket.org/Thinkofdeath/netherrack/system"
	"bitbucket.org/Thinkofdeath/soulsand"
	"bitbucket.org/Thinkofdeath/soulsand/command"
	"bitbucket.org/Thinkofdeath/soulsand/effect"
	sevent "bitbucket.org/Thinkofdeath/soulsand/event"
	"bitbucket.org/Thinkofdeath/soulsand/gamemode"
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	mrand "math/rand"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"
)

var PROTOVERSION byte

//Compile time checks
var _ soulsand.UnsafeConnection = &Connection{}

type Connection struct {
	conn   net.Conn
	player *Player

	outStream cipher.StreamWriter
	inStream  cipher.StreamReader
}

func (c *Connection) GetInputStream() io.Reader {
	return c.inStream
}

func (c *Connection) GetOutputStream() io.Writer {
	return c.outStream
}

var (
	cert   []byte
	priKey *rsa.PrivateKey
)

func init() {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		log.Println(err)
		return
	}
	key.Precompute()

	//cert, err = asn1.Marshal(key.PublicKey)
	cert, err = x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		log.Println(err)
		return
	}
	priKey = key
}

func (c *Connection) Login() {
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	handshake := c.ReadHandshake()
	if handshake.ProtoVersion != PROTOVERSION {
		runtime.Goexit()
	}
	if soulsand.GetServer().GetFlag(soulsand.RANDOM_NAMES) {
		ext := fmt.Sprintf("%d", mrand.Int31n(9999))
		if len(handshake.Username)+len(ext) > 16 {
			handshake.Username = handshake.Username[:16-len(ext)] + ext
		} else {
			handshake.Username += ext
		}
	}
	log.Printf("Player %s connecting\n", handshake.Username)
	c.player.name = handshake.Username
	token := make([]byte, 16)
	rand.Read(token)

	sByte := make([]byte, 4)
	binary.BigEndian.PutUint32(sByte, uint32(mrand.Int()))
	serverID := hex.EncodeToString(sByte)

	c.conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	c.WriteKeyRequest(token, serverID)

	res := make([]byte, 1)
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	_, err := c.conn.Read(res)
	if err != nil {
		log.Println(err)
		runtime.Goexit()
	}
	if res[0] != 0xFC {
		log.Println("Bad response")
		runtime.Goexit()
	}

	response := c.ReadKeyResponse()
	eKey, err := rsa.DecryptPKCS1v15(rand.Reader, priKey, response.Secret)
	if err != nil {
		log.Println(err)
		runtime.Goexit()
	}

	dToken, err := rsa.DecryptPKCS1v15(rand.Reader, priKey, response.VToken)
	if err != nil {
		log.Println(err)
		runtime.Goexit()
	}

	if !bytes.Equal(dToken, token) {
		log.Println("Token mismatch")
		runtime.Goexit()
	}

	//Auth client
	sha := sha1.New()
	sha.Write([]byte(serverID))
	sha.Write([]byte(eKey))
	sha.Write([]byte(cert))
	hash := sha.Sum(make([]byte, 0))
	negative := (hash[0] & 0x80) == 0x80
	if negative {
		hash = twosCompliment(hash)
	}
	buf := hex.EncodeToString(hash)
	if negative {
		buf = "-" + buf
	}

	if !soulsand.GetServer().GetFlag(soulsand.OFFLINE_MODE) {
		hashStr := strings.TrimLeft(buf, "0")
		resp, err := http.Get(fmt.Sprintf("http://session.minecraft.net/game/checkserver.jsp?user=%s&serverId=%s", handshake.Username, hashStr))
		if err != nil {
			log.Println(err)
			runtime.Goexit()
		}
		respB, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Println(err)
			runtime.Goexit()
		}
		if string(respB) != "YES" {
			log.Println("Auth failed")
			runtime.Goexit()
		}
	}
	log.Println("Client auth Ok")

	out := make([]byte, 5)
	out[0] = 0xFC
	c.conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	c.conn.Write(out)

	aCi, err := aes.NewCipher(eKey)
	if err != nil {
		log.Println(err)
		runtime.Goexit()
	}

	c.inStream = cipher.StreamReader{
		R: c.conn,
		S: NewCFB8Decrypt(aCi, eKey),
	}
	c.outStream = cipher.StreamWriter{
		W: c.conn,
		S: NewCFB8Encrypt(aCi, eKey),
	}

	cdPack := make([]byte, 2)
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	c.inStream.Read(cdPack)
	if !bytes.Equal(cdPack, []byte{0xCD, 0x00}) {
		log.Println(cdPack)
		runtime.Goexit()
	}

}

func twosCompliment(p []byte) []byte {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = ^p[i]
		if carry {
			carry = p[i] == 0xFF
			p[i]++
		}
	}
	return p
}

var packets map[byte]func(c *Connection) = map[byte]func(c *Connection){
	0x00: func(c *Connection) { //Keep Alive
		id := c.ReadInt()
		if id != c.player.currentTickID {
			runtime.Goexit()
		}
	},
	0x03: func(c *Connection) { //Chat Message
		msg := c.ReadString()
		if len(msg) <= 0 {
			return
		}
		if msg[0] == '/' {
			command.Exec(msg[1:], c.player)
		} else {
			ev := event.NewMessage(c.player, msg)
			if !c.player.Fire(sevent.PLAYER_MESSAGE, ev) {
				system.Broadcast(ev.GetMessage())
			}
		}
	},
	0x07: func(c *Connection) { //Use Entity
		c.ReadUseEntity()
	},
	0x0A: func(c *Connection) { //Player
		onGround := c.ReadUByte()
		_ = onGround
	},
	0x0B: func(c *Connection) { //Player Position
		pack := c.ReadPlayerPosition()
		if !c.player.IgnoreMoveUpdates {
			c.player.Position.X = pack.X
			c.player.Position.Y = pack.Y
			c.player.Position.Z = pack.Z
		}
	},
	0x0C: func(c *Connection) { //Player Look
		pack := c.ReadPlayerLook()
		c.player.Position.Yaw = pack.Yaw
		c.player.Position.Pitch = pack.Pitch
	},
	0x0D: func(c *Connection) { //Player Position and Look
		pack := c.ReadPlayerPositionLook()
		if !c.player.IgnoreMoveUpdates {
			c.player.Position.X = pack.X
			c.player.Position.Y = pack.Y
			c.player.Position.Z = pack.Z
		}
		c.player.Position.Yaw = pack.Yaw
		c.player.Position.Pitch = pack.Pitch
	},
	0x0E: func(c *Connection) { //Player Digging
		pack := c.ReadPlayerDigging()
		if pack.Status != 2 && !(pack.Status == 0 && c.player.gamemode == gamemode.Creative) {
			return
		}
		x := int(pack.X)
		y := int(pack.Y)
		z := int(pack.Z)

		c.player.World.RunSync(x>>4, z>>4, func(ch soulsand.SyncChunk) {
			chunk := ch.(interface {
				GetPlayerMap() map[int32]soulsand.Player
			})
			rx := x - ((x >> 4) << 4)
			rz := z - ((z >> 4) << 4)
			block := ch.GetBlock(rx, y, rz)
			meta := ch.GetMeta(rx, y, rz)
			m := chunk.GetPlayerMap()
			for _, p := range m {
				if p.GetName() != c.player.GetName() {
					p.PlayEffect(x, y, z, effect.BlockBreak, int(block)|(int(meta)<<12), true)
				}
			}
		})
		c.player.World.SetBlock(x, y, z, 0, 0)
	},
	0x0F: func(c *Connection) { //Player Block Placement
		pack := c.ReadPlayerBlockPlacement()
		x := int(pack.X)
		y := int(pack.Y)
		z := int(pack.Z)
		switch pack.Direction {
		case 0:
			y--
		case 1:
			y++
		case 2:
			z--
		case 3:
			z++
		case 4:
			x--
		case 5:
			x++
		}
		c.player.World.SetBlock(x, y, z, 1, 0)
	},
	0x10: func(c *Connection) { //Held Item Change
		slotID := c.ReadShort()
		c.player.Inventory.CurrentSlot = int(slotID)
	},
	0x12: func(c *Connection) { //Animation
		eID := c.ReadInt()
		ani := c.ReadByte()
		_ = eID
		_ = ani
		/*chunk.SendChunkMessage(c.player.chunk.X, c.player.chunk.Z, &playerAnimation{
			EID: eID,
			Ani: ani,
		})*/
	},
	0x13: func(c *Connection) { //Entity Action
		eID := c.ReadInt()
		aID := c.ReadByte()
		_ = eID
		_ = aID
	},
	0x65: func(c *Connection) { //Close Window
		wID := c.ReadUByte()
		_ = wID
	},
	0x66: func(c *Connection) { //Click Window
		c.ReadClickWindow()
	},
	0x6A: func(c *Connection) { //Confirm Transaction
		c.ReadConfirmTransaction()
	},
	0x6B: func(c *Connection) { //Creative Inventory Action
		c.ReadCreativeInventoryAction()
	},
	0x6C: func(c *Connection) { //Enchant Item
		wID := c.ReadUByte()
		enchantment := c.ReadByte()
		_ = wID
		_ = enchantment
	},
	0x82: func(c *Connection) { //Update Sign
		c.ReadUpdateSign()
	},
	0xCA: func(c *Connection) { //Player Abilities
		flags := c.ReadUByte()
		fSpeed := c.ReadUByte()
		wSpeed := c.ReadUByte()
		_ = flags
		_ = fSpeed
		_ = wSpeed
	},
	0xCB: func(c *Connection) { //Tab-complete
		text := c.ReadString()
		c.WriteTabComplete(command.Complete(text[1:]))
	},
	0xCC: func(c *Connection) { //Client Settings
		settings := c.ReadClientSettings()
		c.player.settings.locale = settings.Locale
		old := c.player.settings.viewDistance
		c.player.settings.viewDistance = int(math.Pow(2, 4-float64(settings.ViewDistance)))
		if c.player.settings.viewDistance > 10 {
			c.player.settings.viewDistance = 10
		}
		if old != c.player.settings.viewDistance {
			c.player.chunkReload(old)
		}
		c.player.settings.chatFlags = settings.ChatFlags
		c.player.settings.difficulty = settings.Difficulty
		c.player.settings.showCape = settings.ShowCape
	},
	0xCD: func(c *Connection) { //Client Statuses
		payload := c.ReadUByte()
		_ = payload
	},
	0xFA: func(c *Connection) { //Plugin Message
		channel := c.ReadString()
		l := c.ReadShort()
		data := make([]byte, l)
		c.inStream.Read(data)
		_ = channel
		_ = data
	},
	0xFF: func(c *Connection) { //Disconnect
		reason := c.ReadString()
		_ = reason
		runtime.Goexit()
	},
}

func (c *Connection) WriteDisconnect(reason string) {
	reasonR := []rune(reason)
	out := make([]byte, 1+2+len(reasonR)*2)
	out[0] = 0xFF
	WriteString(out[1:], reasonR)
	c.Write(out)
}

func (c *Connection) WritePluginMessage(channel string, data []byte) {
	channelR := []rune(channel)
	out := make([]byte, 1+2+len(channelR)*2+2+len(data))
	out[0] = 0xFA
	pos := WriteString(out[1:], channelR) + 1
	WriteShort(out[pos:pos+2], int16(len(data)))
	copy(out[pos+2:], data)
	c.Write(out)
}

func (c *Connection) WriteTeamRemovePlayers(name string, players []string) {
	nameR := []rune(name)
	out := make([]byte, 1+2+len(nameR)*2+1+2)
	out[0] = 0xD1
	pos := WriteString(out[1:], nameR) + 1
	out[pos] = 4
	pos++
	WriteShort(out[pos:], int16(len(players)))
	for _, p := range players {
		pR := []rune(p)
		t := make([]byte, 2+len(pR)*2)
		WriteString(t, pR)
		out = append(out, t...)
	}
	c.outStream.Write(out)
}

func (c *Connection) WriteTeamAddPlayers(name string, players []string) {
	nameR := []rune(name)
	out := make([]byte, 1+2+len(nameR)*2+1+2)
	out[0] = 0xD1
	pos := WriteString(out[1:], nameR) + 1
	out[pos] = 3
	pos++
	WriteShort(out[pos:], int16(len(players)))
	for _, p := range players {
		pR := []rune(p)
		t := make([]byte, 2+len(pR)*2)
		WriteString(t, pR)
		out = append(out, t...)
	}
	c.outStream.Write(out)
}

func (c *Connection) WriteUpdateTeam(name string, dName string, pre string, suf string, ff int8) {
	nameR := []rune(name)
	dNameR := []rune(dName)
	preR := []rune(pre)
	sufR := []rune(suf)
	out := make([]byte, 1+2+len(nameR)*2+1+2+len(dName)*2+2+len(preR)*2+2+len(sufR)*2+1)
	out[0] = 0xD1
	pos := WriteString(out[1:], nameR) + 1
	out[pos] = 0
	pos++
	pos += WriteString(out[pos:], dNameR)
	pos += WriteString(out[pos:], preR)
	pos += WriteString(out[pos:], sufR)
	WriteByte(out[pos:], ff)
	c.outStream.Write(out)
}

func (c *Connection) WriteRemoveTeam(name string) {
	nameR := []rune(name)
	out := make([]byte, 1+2+len(nameR)*2+1)
	out[0] = 0xD1
	pos := WriteString(out[1:], nameR) + 1
	out[pos] = 1
	c.outStream.Write(out)
}

func (c *Connection) WriteCreateTeam(name string, dName string, pre string, suf string, ff bool, players []string) {
	nameR := []rune(name)
	dNameR := []rune(dName)
	preR := []rune(pre)
	sufR := []rune(suf)
	out := make([]byte, 1+2+len(nameR)*2+1+2+len(dName)*2+2+len(preR)*2+2+len(sufR)*2+1+2)
	out[0] = 0xD1
	pos := WriteString(out[1:], nameR) + 1
	out[pos] = 0
	pos++
	pos += WriteString(out[pos:], dNameR)
	pos += WriteString(out[pos:], preR)
	pos += WriteString(out[pos:], sufR)
	WriteBool(out[pos:], ff)
	pos++
	WriteShort(out[pos:], int16(len(players)))
	for _, p := range players {
		pR := []rune(p)
		t := make([]byte, 2+len(pR)*2)
		WriteString(t, pR)
		out = append(out, t...)
	}
	c.outStream.Write(out)
}

func (c *Connection) WriteDisplayScoreboard(position int8, name string) {
	nameR := []rune(name)
	out := make([]byte, 1+2+len(nameR)*2+1)
	out[0] = 0xD0
	WriteByte(out[1:], position)
	WriteString(out[2:], nameR)
	c.Write(out)
}

func (c *Connection) WriteUpdateScore(name string, sName string, value int32) {
	nameR := []rune(name)
	sNameR := []rune(sName)
	out := make([]byte, 1+2+len(nameR)*2+1+2+len(sNameR)*2+4)
	out[0] = 0xCF
	pos := WriteString(out[1:], nameR) + 1
	WriteByte(out[pos:], 0)
	pos++
	pos += WriteString(out[pos:], sNameR)
	WriteInt(out[pos:], value)
	c.Write(out)
}

func (c *Connection) WriteRemoveScore(name string) {
	nameR := []rune(name)
	out := make([]byte, 1+2+len(nameR)*2+1)
	out[0] = 0xCF
	pos := WriteString(out[1:], nameR) + 1
	WriteByte(out[pos:], 1)
	c.Write(out)
}

func (c *Connection) WriteCreateScoreboard(name, display string, remove bool) {
	nameR := []rune(name)
	displayR := []rune(display)
	out := make([]byte, 1+2+len(nameR)*2+2+len(displayR)*2+1)
	out[0] = 0xCE
	pos := WriteString(out[1:], nameR) + 1
	pos += WriteString(out[pos:], displayR)
	WriteBool(out[pos:], remove)
	c.Write(out)
}

func (c *Connection) ReadClientSettings() *ClientSettingsData {
	pack := &ClientSettingsData{}
	pack.Locale = c.ReadString()
	pack.ViewDistance = c.ReadUByte()
	pack.ChatFlags = c.ReadUByte()
	pack.Difficulty = c.ReadUByte()
	if c.ReadUByte() == 1 {
		pack.ShowCape = true
	} else {
		pack.ShowCape = false
	}
	return pack
}

type ClientSettingsData struct {
	Locale       string
	ViewDistance byte
	ChatFlags    byte
	Difficulty   byte
	ShowCape     bool
}

func (c *Connection) WriteTabComplete(text string) {
	textR := []rune(text)
	out := make([]byte, 3+len(textR)*2)
	out[0] = 0xCB
	WriteString(out[1:], textR)
	c.Write(out)
}

func (c *Connection) WritePlayerAbilities(flags, fSpeed, wSpeed byte) {
	out := make([]byte, 4)
	out[0] = 0xCA
	out[1] = flags
	out[2] = fSpeed
	out[3] = wSpeed
	c.Write(out)
}

func (c *Connection) WritePlayerListItem(name string, online bool, ping int16) {
	nameR := []rune(name)
	out := make([]byte, 6+len(nameR)*2)
	out[0] = 0xC9
	pos := WriteString(out[1:], nameR) + 1
	WriteBool(out[pos:], online)
	WriteShort(out[pos+1:], ping)
	c.Write(out)
}

func (c *Connection) WriteIncrementStatistic(statID int32, amount byte) {
	out := make([]byte, 6)
	out[0] = 0xC8
	WriteInt(out[1:5], statID)
	out[5] = amount
	c.Write(out)
}

func (c *Connection) WriteUpdateTileEntity(x int32, y int16, z int32, action byte, data *nbt.Compound) {
	var buf bytes.Buffer
	if data != nil {
		gz, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
		data.WriteTo(gz, true)
		gz.Close()
	}
	out := make([]byte, 14+buf.Len())
	out[0] = 0x84
	WriteInt(out[1:5], x)
	WriteShort(out[5:7], y)
	WriteInt(out[7:11], z)
	out[11] = action
	WriteShort(out[12:14], int16(buf.Len()))
	copy(out[14:], buf.Bytes())
	c.Write(out)
}

func (c *Connection) WriteItemData(iType, iID int16, data []byte) {
	out := make([]byte, 7+len(data))
	out[0] = 0x83
	WriteShort(out[1:3], iType)
	WriteShort(out[3:5], iID)
	WriteShort(out[5:7], int16(len(data)))
	copy(out[7:], data)
	c.Write(out)
}

func (c *Connection) ReadUpdateSign() *UpdateSignData {
	pack := &UpdateSignData{}
	pack.X = c.ReadInt()
	pack.Y = c.ReadShort()
	pack.Z = c.ReadInt()
	pack.Text1 = c.ReadString()
	pack.Text2 = c.ReadString()
	pack.Text3 = c.ReadString()
	pack.Text4 = c.ReadString()
	return pack
}

type UpdateSignData struct {
	X     int32
	Y     int16
	Z     int32
	Text1 string
	Text2 string
	Text3 string
	Text4 string
}

func (c *Connection) WriteUpdateSign(x int32, y int16, z int32, text1, text2, text3, text4 string) {
	text1R := []rune(text1)
	text2R := []rune(text2)
	text3R := []rune(text3)
	text4R := []rune(text4)
	out := make([]byte, 1+4+2+4+2+len(text1R)*2+2+len(text2R)*2+2+len(text3R)*2+2+len(text4R)*2)
	out[0] = 0x82
	WriteInt(out[1:5], x)
	WriteShort(out[5:7], y)
	WriteInt(out[7:11], z)
	pos := WriteString(out[11:], text1R) + 11
	pos += WriteString(out[pos:], text2R)
	pos += WriteString(out[pos:], text3R)
	WriteString(out[pos:], text4R)
	c.Write(out)
}

func (c *Connection) ReadCreativeInventoryAction() *CreativeInventoryActionData {
	pack := &CreativeInventoryActionData{}
	pack.Slot = c.ReadShort()
	item := &items.ItemStack{}
	item.ID = c.ReadShort()
	if item.ID != -1 {
		item.Count = c.ReadUByte()
		item.Damage = c.ReadShort()
		l := c.ReadShort()
		if l != -1 {
			data := make([]byte, l)
			c.inStream.Read(data)
			buf := bytes.NewReader(data)
			gz, err := gzip.NewReader(buf)
			if err != nil {
				runtime.Goexit()
			}
			item.Tag = nbt.ParseCompound(nbt.Reader{R: gz})
			gz.Close()
		}
	}
	pack.ClickedItem = item
	return pack
}

type CreativeInventoryActionData struct {
	Slot        int16
	ClickedItem soulsand.ItemStack
}

func (c *Connection) WriteCreativeInventoryAction(slot int16, soulItem soulsand.ItemStack) {
	out := make([]byte, 3)
	out[0] = 0x6B
	WriteShort(out[1:3], slot)
	c.Write(out)
	item := soulItem.(*items.ItemStack)
	if item.ID == -1 {
		out := make([]byte, 2)
		WriteShort(out, item.ID)
		c.Write(out)
	} else {
		if item.Tag == nil {
			out := make([]byte, 7)
			WriteShort(out[0:2], item.ID)
			out[2] = item.Count
			WriteShort(out[3:5], item.Damage)
			c.Write(out)
		} else {
			out := make([]byte, 7)
			WriteShort(out[0:2], item.ID)
			out[2] = item.Count
			WriteShort(out[3:5], item.Damage)
			var buf bytes.Buffer
			gz, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
			item.Tag.WriteTo(gz, true)
			gz.Close()
			WriteShort(out[5:7], int16(buf.Len()))
			out = append(out, buf.Bytes()...)
			c.Write(out)
		}
	}
}

func (c *Connection) ReadConfirmTransaction() *ConfirmTransactionData {
	pack := &ConfirmTransactionData{}
	pack.WID = c.ReadUByte()
	pack.ANum = c.ReadShort()
	if c.ReadUByte() == 1 {
		pack.Accepted = true
	} else {
		pack.Accepted = false
	}
	return pack
}

type ConfirmTransactionData struct {
	WID      byte
	ANum     int16
	Accepted bool
}

func (c *Connection) WriteConfirmTransaction(wID byte, aNum int16, accepted bool) {
	out := make([]byte, 5)
	out[0] = 0x6A
	out[1] = wID
	WriteShort(out[2:4], aNum)
	WriteBool(out[4:5], accepted)
	c.Write(out)
}

func (c *Connection) WriteUpdateWindowProperty(wID byte, prop, val int16) {
	out := make([]byte, 6)
	out[0] = 0x69
	out[1] = wID
	WriteShort(out[2:4], prop)
	WriteShort(out[4:6], val)
	c.Write(out)
}

func (c *Connection) WriteSetWindowItems(wID byte, slots []soulsand.ItemStack) {
	out := make([]byte, 4)
	out[0] = 0x68
	out[1] = wID
	WriteShort(out[2:4], int16(len(slots)))
	c.Write(out)
	for _, soulSlot := range slots {
		slot := soulSlot.(*items.ItemStack)
		if slot.ID == -1 {
			out := make([]byte, 2)
			WriteShort(out, slot.ID)
			c.Write(out)
		} else {
			if slot.Tag == nil {
				out := make([]byte, 7)
				WriteShort(out[0:2], slot.ID)
				out[2] = slot.Count
				WriteShort(out[3:5], slot.Damage)
				c.Write(out)
			} else {
				out := make([]byte, 7)
				WriteShort(out[0:2], slot.ID)
				out[2] = slot.Count
				WriteShort(out[3:5], slot.Damage)
				var buf bytes.Buffer
				gz, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
				slot.Tag.WriteTo(gz, true)
				gz.Close()
				WriteShort(out[5:7], int16(buf.Len()))
				out = append(out, buf.Bytes()...)
				c.Write(out)
			}
		}
	}
}

func (c *Connection) WriteSetSlot(wID byte, slot int16, soulData soulsand.ItemStack) {
	var out []byte
	data := soulData.(*items.ItemStack)
	if data.ID == -1 {
		out = make([]byte, 6)
		WriteShort(out[4:6], data.ID)
	} else {
		if data.Tag == nil {
			out = make([]byte, 11)
			WriteShort(out[4:6], data.ID)
			out[6] = data.Count
			WriteShort(out[7:9], data.Damage)
			WriteShort(out[9:11], int16(-1))
		} else {
			out = make([]byte, 11)
			WriteShort(out[4:6], data.ID)
			out[6] = data.Count
			WriteShort(out[7:9], data.Damage)
			var buf bytes.Buffer
			gz, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
			data.Tag.WriteTo(gz, true)
			gz.Close()
			WriteShort(out[9:11], int16(buf.Len()))
			out = append(out, buf.Bytes()...)
		}
	}
	out[0] = 0x67
	out[1] = wID
	WriteShort(out[2:4], slot)
	c.Write(out)
}

func (c *Connection) ReadClickWindow() *ClickWindowData {
	pack := &ClickWindowData{}
	pack.WID = c.ReadUByte()
	pack.Slot = c.ReadShort()
	pack.MButton = c.ReadUByte()
	pack.ActionNum = c.ReadShort()
	if c.ReadUByte() == 1 {
		pack.Shift = true
	} else {
		pack.Shift = false
	}
	item := &items.ItemStack{}
	item.ID = c.ReadShort()
	if item.ID != -1 {
		item.Count = c.ReadUByte()
		item.Damage = c.ReadShort()
		l := c.ReadShort()
		if l != -1 {
			data := make([]byte, l)
			c.inStream.Read(data)
			buf := bytes.NewReader(data)
			gz, err := gzip.NewReader(buf)
			if err != nil {
				runtime.Goexit()
			}
			item.Tag = nbt.ParseCompound(nbt.Reader{R: gz})
			gz.Close()
		}
	}
	pack.ClickedItem = item
	return pack
}

type ClickWindowData struct {
	WID         byte
	Slot        int16
	MButton     byte
	ActionNum   int16
	Shift       bool
	ClickedItem soulsand.ItemStack
}

func (c *Connection) WriteCloseWindow(wID byte) {
	out := make([]byte, 2)
	out[0] = 0x65
	out[1] = wID
	c.Write(out)
}

func (c *Connection) WriteOpenWindow(wID, iType byte, title string, slots byte, useTitle bool) {
	titleR := []rune(title)
	out := make([]byte, 7+len(titleR)*2)
	out[0] = 0x64
	out[1] = wID
	out[2] = iType
	pos := WriteString(out[3:], titleR)
	pos += 3
	out[pos] = slots
	WriteBool(out[pos+1:], useTitle)
	c.Write(out)
}

func (c *Connection) WriteSpawnGlobalEntity(eID int32, t byte, x, y, z int32) {
	out := make([]byte, 18)
	out[0] = 0x47
	WriteInt(out[1:5], eID)
	out[5] = t
	WriteInt(out[6:10], x)
	WriteInt(out[10:14], y)
	WriteInt(out[14:18], z)
	c.Write(out)
}

func (c *Connection) WriteChangeGameState(res, gMode byte) {
	out := make([]byte, 3)
	out[0] = 0x46
	out[1] = res
	out[2] = gMode
	c.Write(out)
}

func (c *Connection) WriteParticle(particleName string, x, y, z, ox, oy, oz, speed float32, count int32) {
	particleNameR := []rune(particleName)
	out := make([]byte, 1+2+len(particleNameR)*2+4+4+4+4+4+4+4+4)
	out[0] = 0x3F
	pos := WriteString(out[1:], particleNameR) + 1
	WriteFloat(out[pos:], x)
	pos += 4
	WriteFloat(out[pos:], y)
	pos += 4
	WriteFloat(out[pos:], z)
	pos += 4
	WriteFloat(out[pos:], ox)
	pos += 4
	WriteFloat(out[pos:], oy)
	pos += 4
	WriteFloat(out[pos:], oz)
	pos += 4
	WriteFloat(out[pos:], speed)
	pos += 4
	WriteInt(out[pos:], count)
	c.outStream.Write(out)
}

func (c *Connection) WriteNameSoundEffect(name string, posX, posY, posZ int32, vol float32, pitch byte) {
	nameR := []rune(name)
	out := make([]byte, 20+len(nameR)*2)
	out[0] = 0x3E
	pos := WriteString(out[1:], nameR)
	pos++
	WriteInt(out[pos:pos+4], posX)
	WriteInt(out[pos+4:pos+8], posY)
	WriteInt(out[pos+8:pos+12], posZ)
	WriteFloat(out[pos+12:pos+16], vol)
	out[pos+16] = pitch
	c.Write(out)
}

func (c *Connection) WriteSoundParticleEffect(eff, x int32, y byte, z, data int32, rVol bool) {
	out := make([]byte, 19)
	out[0] = 0x3D
	WriteInt(out[1:5], eff)
	WriteInt(out[5:9], x)
	out[9] = y
	WriteInt(out[10:14], z)
	WriteInt(out[14:18], data)
	WriteBool(out[18:19], rVol)
	c.Write(out)
}

func (c *Connection) WriteExplosion(x, y, z float64, radius float32, records []byte, velX, velY, velZ float32) {
	out := make([]byte, 1+8+8+8+4+4+len(records)+4+4+4)
	out[0] = 0x3C
	WriteDouble(out[1:9], x)
	WriteDouble(out[9:17], y)
	WriteDouble(out[17:25], z)
	WriteFloat(out[25:29], radius)
	WriteInt(out[29:33], int32(len(records)/3))
	copy(out[33:], records)
	WriteFloat(out[33+len(records):], velX)
	WriteFloat(out[33+len(records)+4:], velY)
	WriteFloat(out[33+len(records)+8:], velZ)
	c.Write(out)
}

func (c *Connection) WriteMapChunkBulk() {
	//TODO?
	panic("Map Chunk Bulk(0x38) Not implemented")
}

func (c *Connection) WriteBlockBreakAnimation(eID, x, y, z int32, stage byte) {
	out := make([]byte, 18)
	out[0] = 0x37
	WriteInt(out[1:5], eID)
	WriteInt(out[5:9], x)
	WriteInt(out[9:13], y)
	WriteInt(out[13:17], z)
	out[17] = stage
	c.Write(out)
}

func (c *Connection) WriteBlockAction(x int32, y int16, z int32, b1, b2 byte, bID int16) {
	out := make([]byte, 15)
	out[0] = 0x36
	WriteInt(out[1:5], x)
	WriteShort(out[5:7], y)
	WriteInt(out[7:11], z)
	out[11] = b1
	out[12] = b2
	WriteShort(out[13:15], bID)
	c.Write(out)
}

func (c *Connection) WriteBlockChange(x int32, y byte, z int32, bType int16, bMeta byte) {
	out := make([]byte, 13)
	out[0] = 0x35
	WriteInt(out[1:5], x)
	out[5] = y
	WriteInt(out[6:10], z)
	WriteShort(out[10:12], bType)
	out[12] = bMeta
	c.Write(out)
}

func (c *Connection) WriteMultiBlockChange(cx, cz int32, blocks []uint32) {
	out := make([]byte, 15+4*len(blocks))
	out[0] = 0x34
	WriteInt(out[1:5], cx)
	WriteInt(out[5:9], cz)
	WriteShort(out[9:11], int16(len(blocks)))
	WriteInt(out[11:15], int32(len(blocks)*4))
	for i := 0; i < len(blocks); i++ {
		//b := blocks[i]
		//var d uint32 = (uint32(b.Metadata) & 0xf) | (uint32(b.ID) << 4) | (uint32(b.Y) << 16) | (uint32(b.Z) << 24) | (uint32(b.X) << 28)
		binary.BigEndian.PutUint32(out[15+i*4:], blocks[i])
	}
	c.Write(out)
}

func (c *Connection) WriteChunkDataUnload(x, z int32) {
	out := make([]byte, 1+4+4+1+2+2+4)
	out[0] = 0x33
	WriteInt(out[1:5], x)
	WriteInt(out[5:9], z)
	WriteBool(out[9:10], true)
	c.Write(out)
}

type BlockChangeData struct {
	ID       byte
	Metadata byte
	X        byte
	Y        byte
	Z        byte
}

func (c *Connection) WriteSetExperience(bar float32, level, exp int16) {
	out := make([]byte, 9)
	out[0] = 0x2B
	WriteFloat(out[1:5], bar)
	WriteShort(out[5:7], level)
	WriteShort(out[7:9], exp)
	c.Write(out)
}

func (c *Connection) WriteRemoveEntityEffect(eID int32, eff int8) {
	out := make([]byte, 6)
	out[0] = 0x2A
	WriteInt(out[1:5], eID)
	WriteByte(out[5:6], eff)
	c.Write(out)
}

func (c *Connection) WriteEntityEffect(eID int32, eff int8, amp int8, duration int16) {
	out := make([]byte, 9)
	out[0] = 0x29
	WriteInt(out[1:5], eID)
	WriteByte(out[5:6], eff)
	WriteByte(out[6:7], amp)
	WriteShort(out[7:9], duration)
	c.Write(out)
}

func (c *Connection) WriteEntityMetadata(eID int32, metadata map[byte]entity.MetadataItem) {
	out := make([]byte, 5)
	out[0] = 0x28
	WriteInt(out[1:5], eID)
	c.Write(out)
	c.writeMetadata(metadata)
}

func (c *Connection) WriteAttachEntity(eID int32, vID int32) {
	out := make([]byte, 9)
	out[0] = 0x27
	WriteInt(out[1:5], eID)
	WriteInt(out[5:9], vID)
	c.Write(out)
}

func (c *Connection) WriteEntityStatus(eID int32, status int8) {
	out := make([]byte, 6)
	out[0] = 0x26
	WriteInt(out[1:5], eID)
	WriteByte(out[5:6], status)
	c.Write(out)
}

func (c *Connection) WriteEntityHeadLook(eID int32, hYaw int8) {
	out := make([]byte, 6)
	out[0] = 0x23
	WriteInt(out[1:5], eID)
	WriteByte(out[5:6], hYaw)
	c.Write(out)
}

func (c *Connection) WriteEntityTeleport(eID, x, y, z int32, yaw, pitch int8) {
	out := make([]byte, 19)
	out[0] = 0x22
	WriteInt(out[1:5], eID)
	WriteInt(out[5:9], x)
	WriteInt(out[9:13], y)
	WriteInt(out[13:17], z)
	WriteByte(out[17:18], yaw)
	WriteByte(out[18:19], pitch)
	c.Write(out)
}

func (c *Connection) WriteEntityLookRelativeMove(eID int32, dX, dY, dZ int8, yaw, pitch int8) {
	out := make([]byte, 10)
	out[0] = 0x21
	WriteInt(out[1:5], eID)
	WriteByte(out[5:6], dX)
	WriteByte(out[6:7], dY)
	WriteByte(out[7:8], dZ)
	WriteByte(out[8:9], yaw)
	WriteByte(out[9:10], pitch)
	c.Write(out)
}

func (c *Connection) WriteEntityLook(eID int32, yaw, pitch int8) {
	out := make([]byte, 7)
	out[0] = 0x20
	WriteInt(out[1:5], eID)
	WriteByte(out[5:6], yaw)
	WriteByte(out[6:7], pitch)
	c.Write(out)
}

func (c *Connection) WriteEntityRelativeMove(eID int32, dX, dY, dZ int8) {
	out := make([]byte, 8)
	out[0] = 0x1F
	WriteInt(out[1:5], eID)
	WriteByte(out[5:6], dX)
	WriteByte(out[6:7], dY)
	WriteByte(out[7:8], dZ)
	c.Write(out)
}

func (c *Connection) WriteEntity(eID int32) {
	out := make([]byte, 5)
	out[0] = 0x1E
	WriteInt(out[1:5], eID)
	c.Write(out)
}

func (c *Connection) WriteDestroyEntity(eIDS []int32) {
	out := make([]byte, 2+4*len(eIDS))
	out[0] = 0x1D
	out[1] = byte(len(eIDS))
	for i := 0; i < len(eIDS); i++ {
		WriteInt(out[2+i*4:], eIDS[i])
	}
	c.Write(out)
}

func (c *Connection) WriteEntityVelocity(eID int32, velX, velY, velZ int16) {
	out := make([]byte, 11)
	out[0] = 0x1C
	WriteInt(out[1:5], eID)
	WriteShort(out[5:7], velX)
	WriteShort(out[7:9], velY)
	WriteShort(out[9:11], velZ)
	c.Write(out)
}

func (c *Connection) WriteSpawnExperienceOrb(eID, x, y, z int32, count int16) {
	out := make([]byte, 19)
	out[0] = 0x1A
	WriteInt(out[1:5], eID)
	WriteInt(out[5:9], x)
	WriteInt(out[9:13], y)
	WriteInt(out[13:17], z)
	WriteShort(out[17:19], count)
	c.Write(out)
}

func (c *Connection) WriteSpawnPainting(eID int32, title string, x, y, z, dir int32) {
	titleR := []rune(title)
	out := make([]byte, 1+4+2+len(titleR)*2+4+4+4+4)
	out[0] = 0x19
	WriteInt(out[1:5], eID)
	pos := WriteString(out[5:], titleR)
	pos += 5
	WriteInt(out[pos:pos+4], x)
	WriteInt(out[pos+4:pos+8], y)
	WriteInt(out[pos+8:pos+12], z)
	WriteInt(out[pos+12:pos+16], dir)
	c.Write(out)
}

func (c *Connection) WriteSpawnMob(eID int32, t int8, x, y, z int32, yaw, pitch, hYaw int8, velX, velY, velZ int16, metadata map[byte]entity.MetadataItem) {
	out := make([]byte, 27)
	out[0] = 0x18
	WriteInt(out[1:5], eID)
	WriteByte(out[5:6], t)
	WriteInt(out[6:10], x)
	WriteInt(out[10:14], y)
	WriteInt(out[14:18], z)
	WriteByte(out[18:19], yaw)
	WriteByte(out[19:20], pitch)
	WriteByte(out[20:21], hYaw)
	WriteShort(out[21:23], velX)
	WriteShort(out[23:25], velY)
	WriteShort(out[25:27], velZ)
	c.Write(out)
	c.writeMetadata(metadata)
}

func (c *Connection) WriteSpawnObjectSpeed(eID int32, t int8, x, y, z int32, yaw, pitch int8, data int32, speedX, speedY, speedZ int16) {
	out := make([]byte, 30)
	out[0] = 0x17
	WriteInt(out[1:5], eID)
	WriteByte(out[5:6], t)
	WriteInt(out[6:10], x)
	WriteInt(out[10:14], y)
	WriteInt(out[14:18], z)
	WriteByte(out[18:19], yaw)
	WriteByte(out[19:20], pitch)
	WriteInt(out[20:24], data)
	WriteShort(out[24:26], speedX)
	WriteShort(out[26:28], speedY)
	WriteShort(out[28:30], speedZ)
	c.Write(out)
}

func (c *Connection) WriteSpawnObject(eID int32, t int8, x, y, z int32, yaw, pitch int8) {
	out := make([]byte, 24)
	out[0] = 0x17
	WriteInt(out[1:5], eID)
	WriteByte(out[5:6], t)
	WriteInt(out[6:10], x)
	WriteInt(out[10:14], y)
	WriteInt(out[14:18], z)
	WriteByte(out[18:19], yaw)
	WriteByte(out[19:20], pitch)
	WriteInt(out[20:24], 0)
	c.Write(out)
}

func (c *Connection) WriteCollectItem(collectedEID, collectorEID int32) {
	out := make([]byte, 9)
	out[0] = 0x16
	WriteInt(out[1:5], collectedEID)
	WriteInt(out[6:9], collectorEID)
	c.Write(out)
}

func (c *Connection) WriteSpawnNamedEntity(eID int32, name string, x, y, z int32, yaw, pitch int8, curItem int16, metadata map[byte]entity.MetadataItem) {
	nameR := []rune(name)
	out := make([]byte, 1+4+2+len(nameR)*2+4+4+4+1+1+2)
	out[0] = 0x14
	WriteInt(out[1:5], eID)
	pos := WriteString(out[5:], nameR)
	pos += 5
	WriteInt(out[pos:], x)
	WriteInt(out[pos+4:], y)
	WriteInt(out[pos+8:], z)
	WriteByte(out[pos+12:], yaw)
	WriteByte(out[pos+13:], pitch)
	WriteShort(out[pos+14:], curItem)
	c.Write(out)
	c.writeMetadata(metadata)
}

func (c *Connection) writeMetadata(metadata map[byte]entity.MetadataItem) {
	for _, i := range metadata {
		var out []byte
		switch i.Type {
		case 0:
			out = make([]byte, 2)
			data := i.Value.(int8)
			WriteByte(out[1:2], data)
		case 1:
			out = make([]byte, 3)
			data := i.Value.(int16)
			WriteShort(out[1:3], data)
		case 2:
			out = make([]byte, 5)
			data := i.Value.(int32)
			WriteInt(out[1:5], data)
		case 3:
			out = make([]byte, 5)
			data := i.Value.(float32)
			WriteFloat(out[1:5], data)
		case 4:
			data := []rune(i.Value.(string))
			out = make([]byte, 3+len(data)*2)
			WriteString(out[1:], data)
		case 5:
			slot := i.Value.(*items.ItemStack)
			if slot.ID == -1 {
				out = make([]byte, 3)
				WriteShort(out[1:3], slot.ID)
			} else {
				if slot.Tag == nil {
					out = make([]byte, 8)
					WriteShort(out[1:3], slot.ID)
					out[3] = slot.Count
					WriteShort(out[4:6], slot.Damage)
					WriteShort(out[6:8], 0)
				} else {
					out = make([]byte, 8)
					WriteShort(out[1:3], slot.ID)
					out[3] = slot.Count
					WriteShort(out[4:6], slot.Damage)
					var buf bytes.Buffer
					gz, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
					slot.Tag.WriteTo(gz, true)
					gz.Close()
					WriteShort(out[6:8], int16(buf.Len()))
					out = append(out, buf.Bytes()...)
				}
			}
		case 6:
			out = make([]byte, 13)
			data := i.Value.([]int32)
			WriteInt(out[1:5], data[0])
			WriteInt(out[5:9], data[1])
			WriteInt(out[9:13], data[2])
		}
		out[0] = (i.Index & 0x1F) | ((i.Type << 5) & 0xE0)
		c.Write(out)
	}
	c.Write([]byte{127})
}

func (c *Connection) WriteAnimation(eID int32, ani int8) {
	out := make([]byte, 6)
	out[0] = 0x12
	WriteInt(out[1:5], eID)
	WriteByte(out[5:6], ani)
	c.Write(out)
}

func (c *Connection) WriteUseBed(eID int32, x int32, y byte, z int32) {
	out := make([]byte, 15)
	out[0] = 0x11
	WriteInt(out[1:5], eID)
	WriteInt(out[6:10], x)
	out[10] = y
	WriteInt(out[11:15], z)
}

func (c *Connection) WriteHeldItemChange(slotID int16) {
	out := make([]byte, 3)
	out[0] = 0x10
	WriteShort(out[1:3], slotID)
	c.Write(out)
}

func (c *Connection) ReadPlayerBlockPlacement() *PlayerBlockPlacementData {
	pack := &PlayerBlockPlacementData{}
	pack.X = c.ReadInt()
	pack.Y = c.ReadUByte()
	pack.Z = c.ReadInt()
	pack.Direction = c.ReadUByte()

	item := &items.ItemStack{}
	item.ID = c.ReadShort()
	if item.ID != -1 {
		item.Count = c.ReadUByte()
		item.Damage = c.ReadShort()
		l := c.ReadShort()
		if l != -1 {
			data := make([]byte, l)
			c.inStream.Read(data)
			buf := bytes.NewReader(data)
			gz, err := gzip.NewReader(buf)
			if err != nil {
				runtime.Goexit()
			}
			item.Tag = nbt.ParseCompound(nbt.Reader{R: gz})
			gz.Close()
		}
	}
	pack.HeldItem = item
	pack.CursorX = c.ReadUByte()
	pack.CursorY = c.ReadUByte()
	pack.CursorZ = c.ReadUByte()
	return pack
}

type PlayerBlockPlacementData struct {
	X         int32
	Y         byte
	Z         int32
	Direction byte
	HeldItem  soulsand.ItemStack
	CursorX   byte
	CursorY   byte
	CursorZ   byte
}

func (c *Connection) ReadPlayerDigging() *PlayerDiggingData {
	pack := &PlayerDiggingData{}
	pack.Status = c.ReadUByte()
	pack.X = c.ReadInt()
	pack.Y = c.ReadUByte()
	pack.Z = c.ReadInt()
	pack.Face = c.ReadUByte()
	return pack
}

type PlayerDiggingData struct {
	Status byte
	X      int32
	Y      byte
	Z      int32
	Face   byte
}

func (c *Connection) ReadPlayerPositionLook() *PlayerPositionLookData {
	pack := &PlayerPositionLookData{}
	pack.X = c.ReadDouble()
	pack.Y = c.ReadDouble()
	pack.Stance = c.ReadDouble()
	pack.Z = c.ReadDouble()
	pack.Yaw = c.ReadFloat()
	pack.Pitch = c.ReadFloat()
	if c.ReadUByte() == 1 {
		pack.OnGround = true
	} else {
		pack.OnGround = false
	}
	return pack

}

type PlayerPositionLookData struct {
	X        float64
	Y        float64
	Z        float64
	Stance   float64
	Yaw      float32
	Pitch    float32
	OnGround bool
}

func (c *Connection) ReadPlayerLook() *PlayerLookData {
	pack := &PlayerLookData{}
	pack.Yaw = c.ReadFloat()
	pack.Pitch = c.ReadFloat()
	if c.ReadUByte() == 1 {
		pack.OnGround = true
	} else {
		pack.OnGround = false
	}
	return pack
}

type PlayerLookData struct {
	Yaw      float32
	Pitch    float32
	OnGround bool
}

func (c *Connection) ReadPlayerPosition() *PlayerPositionData {
	pack := &PlayerPositionData{}
	pack.X = c.ReadDouble()
	pack.Y = c.ReadDouble()
	pack.Stance = c.ReadDouble()
	pack.Z = c.ReadDouble()
	if c.ReadUByte() == 1 {
		pack.OnGround = true
	} else {
		pack.OnGround = false
	}
	return pack
}

type PlayerPositionData struct {
	X        float64
	Y        float64
	Z        float64
	Stance   float64
	OnGround bool
}

func (c *Connection) WriteRespawn(dim int32, diff, gMode int8, height int16, lType string) {
	lTypeR := []rune(lType)
	out := make([]byte, 1+4+1+1+2+2+len(lTypeR)*2)
	out[0] = 0x09
	WriteInt(out[1:5], dim)
	WriteByte(out[5:6], diff)
	WriteByte(out[6:7], gMode)
	WriteShort(out[7:9], height)
	WriteString(out[9:], lTypeR)
	c.Write(out)
}

func (c *Connection) WriteUpdateHealth(health, food int16, fSat float32) {
	out := make([]byte, 9)
	out[0] = 0x08
	WriteShort(out[1:3], health)
	WriteShort(out[3:5], food)
	WriteFloat(out[5:9], fSat)
	c.Write(out)
}

func (c *Connection) ReadUseEntity() *UseEntityData {
	pack := &UseEntityData{}
	pack.User = c.ReadInt()
	pack.Target = c.ReadInt()
	if c.ReadUByte() == 1 {
		pack.Button = true
	} else {
		pack.Button = false
	}
	return pack
}

type UseEntityData struct {
	User   int32
	Target int32
	Button bool
}

func (c *Connection) WriteEntityEquipment(eID int32, slot int16, soulSlotData soulsand.ItemStack) {
	var out []byte
	slotData := soulSlotData.(*items.ItemStack)
	if slotData.ID != -1 {
		if slotData.Tag != nil {
			var b bytes.Buffer
			gz, _ := gzip.NewWriterLevel(&b, gzip.BestSpeed)
			slotData.Tag.WriteTo(gz, false)
			gz.Close()
			slotB := b.Bytes()
			out = make([]byte, 1+4+2+2+1+2+2+len(slotB))
			copy(out[1+4+2+2+1+2+2:], slotB)
			binary.BigEndian.PutUint16(out[1+4+2+2+1+2:], uint16(len(slotB)))
		} else {
			out = make([]byte, 1+4+2+2+1+2+2)
			WriteShort(out[1+4+2+2+1+2:], -1)
		}
		out[1+4+2+2] = slotData.Count
		binary.BigEndian.PutUint16(out[1+4+2+2+1:], uint16(slotData.Damage))
	} else {
		out = make([]byte, 1+4+2+2)
	}
	out[0] = 0x05
	binary.BigEndian.PutUint16(out[1+4+2:], uint16(slotData.ID))
	binary.BigEndian.PutUint16(out[1+4:], uint16(slot))
	binary.BigEndian.PutUint32(out[1:], uint32(eID))
	c.Write(out)
}

func (c *Connection) WriteTimeUpdate(age, time int64) {
	out := make([]byte, 17)
	out[0] = 0x04
	binary.BigEndian.PutUint64(out[1:9], uint64(age))
	binary.BigEndian.PutUint64(out[9:17], uint64(time))
	c.Write(out)
}

func (c *Connection) WriteChatMessage(msg string) {
	msgR := []rune(msg)
	out := make([]byte, 3+len(msgR)*2)
	out[0] = 0x03
	WriteString(out[1:], msgR)
	c.Write(out)
}

func (c *Connection) WriteKeepAlive(id int32) {
	out := make([]byte, 5)
	out[0] = 0x00
	WriteInt(out[1:5], id)
	c.Write(out)
}

func (c *Connection) WritePlayerPositionLook(x, y, z, stance float64, yaw, pitch float32, onGround bool) {
	out := make([]byte, 42)
	out[0] = 0x0D
	WriteDouble(out[1:9], x)
	WriteDouble(out[9:17], stance)
	WriteDouble(out[17:25], y)
	WriteDouble(out[25:33], z)
	WriteFloat(out[33:37], yaw)
	WriteFloat(out[37:41], pitch)
	WriteBool(out[41:42], onGround)
	c.Write(out)
}

func (c *Connection) WriteSpawnPosition(x, y, z int32) {
	out := make([]byte, 13)
	out[0] = 0x06
	WriteInt(out[1:5], x)
	WriteInt(out[5:9], y)
	WriteInt(out[9:13], z)
	c.Write(out)
}

func (c *Connection) WriteLoginRequest(eID int32, lType string, gMode, dim, diff, mP int8) {
	lTypeR := []rune(lType)
	out := make([]byte, 1+4+2+len(lTypeR)*2+5)
	out[0] = 0x01
	WriteInt(out[1:5], eID)
	pos := WriteString(out[5:], lTypeR)
	pos += 5
	WriteByte(out[pos:pos+1], gMode)
	pos++
	WriteByte(out[pos:pos+1], dim)
	pos++
	WriteByte(out[pos:pos+1], diff)
	pos += 2
	WriteByte(out[pos:pos+1], mP)
	c.Write(out)
}

func (c *Connection) WriteKeyRequest(token []byte, serverID string) {
	serverIDR := []rune(serverID)
	out := make([]byte, 1+2+len(serverIDR)*2+2+len(cert)+2+len(token))
	out[0] = 0xFD
	l := WriteString(out[1:], serverIDR)
	pos := 1 + l
	WriteShort(out[pos:pos+2], int16(len(cert)))
	pos += 2
	copy(out[pos:pos+len(cert)], cert)
	pos += len(cert)
	WriteShort(out[pos:pos+2], int16(len(token)))
	pos += 2
	copy(out[pos:pos+len(token)], token)
	c.conn.Write(out)
}

func (c *Connection) ReadKeyResponse() *packetKeyResponse {
	pack := packetKeyResponse{}
	sLen := c.rawReadShort()
	pack.Secret = make([]byte, sLen)
	c.conn.Read(pack.Secret)

	tLen := c.rawReadShort()
	pack.VToken = make([]byte, tLen)
	c.conn.Read(pack.VToken)

	return &pack
}

func (c *Connection) ReadHandshake() *packetHandshake {
	pack := packetHandshake{}
	pack.ProtoVersion = c.rawReadByte()
	pack.Username = c.rawReadString()
	pack.Host = c.rawReadString()
	pack.Port = int(c.rawReadInt())
	return &pack
}

func (c *Connection) Write(out []byte) {
	c.conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	n, err := c.outStream.Write(out)
	if n != len(out) || err != nil {
		runtime.Goexit()
	}
}

func (c *Connection) ReadString() string {
	l := int(c.ReadShort())
	data := make([]byte, l*2)
	r := make([]rune, l)
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	n, err := c.inStream.Read(data)
	if n != l*2 || err != nil {
		runtime.Goexit()
	}
	for i := 0; i < l; i++ {
		r[i] = rune(binary.BigEndian.Uint16(data[i*2 : i*2+2]))
	}
	return string(r)
}

func (c *Connection) ReadDouble() float64 {
	b := make([]byte, 8)
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	n, err := c.inStream.Read(b)
	if n != 8 || err != nil {
		runtime.Goexit()
	}
	return math.Float64frombits(binary.BigEndian.Uint64(b))
}

func (c *Connection) ReadFloat() float32 {
	b := make([]byte, 4)
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	n, err := c.inStream.Read(b)
	if n != 4 || err != nil {
		runtime.Goexit()
	}
	return math.Float32frombits(binary.BigEndian.Uint32(b))
}

func (c *Connection) ReadLong() int64 {
	b := make([]byte, 8)
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	n, err := c.inStream.Read(b)
	if n != 8 || err != nil {
		runtime.Goexit()
	}
	return int64(binary.BigEndian.Uint64(b))
}

func (c *Connection) ReadInt() int32 {
	b := make([]byte, 4)
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	n, err := c.inStream.Read(b)
	if n != 4 || err != nil {
		runtime.Goexit()
	}
	return int32(binary.BigEndian.Uint32(b))
}

func (c *Connection) ReadShort() int16 {
	b := make([]byte, 2)
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	n, err := c.inStream.Read(b)
	if n != 2 || err != nil {
		runtime.Goexit()
	}
	return int16(binary.BigEndian.Uint16(b))
}

func (c *Connection) ReadUByte() byte {
	b := make([]byte, 1)
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	n, err := c.inStream.Read(b)
	if n != 1 || err != nil {
		runtime.Goexit()
	}
	return b[0]
}

func (c *Connection) ReadByte() int8 {
	b := make([]byte, 1)
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	n, err := c.inStream.Read(b)
	if n != 1 || err != nil {
		runtime.Goexit()
	}
	return int8(b[0])
}

func WriteBool(out []byte, b bool) {
	if b {
		out[0] = 1
	} else {
		out[0] = 0
	}
}

func WriteByte(out []byte, b int8) {
	out[0] = byte(b)
}

func WriteShort(out []byte, s int16) {
	binary.BigEndian.PutUint16(out, uint16(s))
}

func WriteUShort(out []byte, s uint16) {
	binary.BigEndian.PutUint16(out, s)
}

func WriteInt(out []byte, s int32) {
	binary.BigEndian.PutUint32(out, uint32(s))
}

func WriteFloat(out []byte, s float32) {
	binary.BigEndian.PutUint32(out, math.Float32bits(s))
}

func WriteDouble(out []byte, s float64) {
	binary.BigEndian.PutUint64(out, math.Float64bits(s))
}

func WriteString(out []byte, r []rune) int {
	//r := []rune(str)
	WriteShort(out[0:2], int16(len(r)))
	pos := 2
	for _, b := range r {
		binary.BigEndian.PutUint16(out[pos:pos+2], uint16(b))
		pos += 2
	}
	return 2 + len(r)*2
}

func (c *Connection) rawReadByte() byte {
	b := make([]byte, 1)
	n, err := c.conn.Read(b)
	if n != 1 || err != nil {
		log.Println(err)
		runtime.Goexit()
	}
	return b[0]
}

func (c *Connection) rawReadString() string {
	length := int(c.rawReadShort())
	r := make([]rune, length)
	for i := 0; i < length; i++ {
		r[i] = rune(c.rawReadShort())
	}
	return string(r)
}

func (c *Connection) rawReadShort() int16 {
	b := make([]byte, 2)
	n, err := c.conn.Read(b)
	if n != 2 || err != nil {
		log.Println(err)
		runtime.Goexit()
	}
	return int16(binary.BigEndian.Uint16(b))
}

func (c *Connection) rawReadInt() int32 {
	b := make([]byte, 4)
	n, err := c.conn.Read(b)
	if n != 4 || err != nil {
		log.Println(err)
		runtime.Goexit()
	}
	return int32(binary.BigEndian.Uint32(b))
}

type (
	packetHandshake struct {
		ProtoVersion byte
		Username     string
		Host         string
		Port         int
	}

	packetKeyResponse struct {
		Secret []byte
		VToken []byte
	}
)

/*
	Allow for AES streams
*/
type CFB8 struct {
	c         cipher.Block
	blockSize int
	iv        []byte
	tmp       []byte
	de        bool
}

func NewCFB8Decrypt(c cipher.Block, iv []byte) *CFB8 {
	cp := make([]byte, len(iv))
	copy(cp, iv)
	return &CFB8{
		c:         c,
		blockSize: c.BlockSize(),
		iv:        cp,
		tmp:       make([]byte, c.BlockSize()),
		de:        true,
	}
}

func NewCFB8Encrypt(c cipher.Block, iv []byte) *CFB8 {
	cp := make([]byte, len(iv))
	copy(cp, iv)
	return &CFB8{
		c:         c,
		blockSize: c.BlockSize(),
		iv:        cp,
		tmp:       make([]byte, c.BlockSize()),
		de:        false,
	}
}

func (cf *CFB8) XORKeyStream(dst, src []byte) {
	for i := 0; i < len(src); i++ {
		val := src[i]
		copy(cf.tmp, cf.iv)
		cf.c.Encrypt(cf.iv, cf.iv)
		val = val ^ cf.iv[0]

		copy(cf.iv, cf.tmp[1:])
		if cf.de {
			cf.iv[15] = src[i]
		} else {
			cf.iv[15] = val
		}

		dst[i] = val
	}
}
