package chunk

import ()

type lightInfo struct {
	x     int
	y     int
	z     int
	light byte
	next  *lightInfo
	root  *lightInfo
}

func (l *lightInfo) Append(l2 *lightInfo) *lightInfo {
	if l != nil {
		l.next = l2
		l2.root = l.root
	} else {
		l2.root = l2
	}
	return l2
}

func (l *lightInfo) Free() {
	l.root = nil
	l.next = nil
	//lightInfoFree <- l
}

func LightInfoGet(x, y, z int, light byte) *lightInfo {
	/*select {
	case info := <-lightInfoGet:
		info.x = x
		info.y = y
		info.z = z
		info.light = light
		return info
	default:*/
	return &lightInfo{x: x, y: y, z: z, light: light}
	//}

}

//Light info pooling is disabled because the speed cost is too great currently

// var (
// 	lightInfoGet  = make(chan *lightInfo, 100000) //Storage for lightInfos
// 	lightInfoFree = make(chan *lightInfo, 2000)
// )

// func init() {
// 	go lightInfoPool()
// }

// func lightInfoPool() {
// 	for i := 0; i < cap(lightInfoGet); i++ {
// 		lightInfoGet <- &lightInfo{}
// 	}
// 	for {
// 		select {
// 		case info := <-lightInfoFree:
// 			select {
// 			case lightInfoGet <- info:
// 			default:
// 			}
// 		}
// 	}
// }
