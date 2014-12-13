package overlay

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/henrikssn/stored/route"
	"github.com/henrikssn/stored/store"
	"github.com/stathat/consistent"
	"hash/crc32"
	"log"
	"net"
	"strconv"
	"time"
)

var _ = proto.Marshal

type Overlay struct {
	router      *route.Router
	rpcHandler  *RpcHandler
	store       *store.Store
	me          *Node
	coordinator *Node
	hash        *consistent.Consistent
	members     map[int64]*Node
	nextId      int64
}

type Node struct {
	addr      *net.UDPAddr
	id        int64
	lastAlive time.Time
}

type Location struct {
	id       int64
	replicas []int64
}

func (n *Node) toMember() *Member {
	addr := n.addr.String()
	return &Member{
		Id:      &n.id,
		UdpAddr: &addr,
	}
}

func NodeFrom(member *Member) *Node {
	uaddr, err := net.ResolveUDPAddr("udp", member.GetUdpAddr())
	if err != nil {
		panic(err)
	}
	return &Node{addr: uaddr, id: member.GetId()}
}

func (n *Node) SetAlive() {
	n.lastAlive = time.Now()
}

func New(r *route.Router, s *store.Store, laddr *net.UDPAddr, caddr *net.UDPAddr) *Overlay {
	o := new(Overlay)
	o.router = r
	o.me = &Node{addr: laddr, id: int64(1)}
	o.coordinator = &Node{addr: caddr}
	o.members = make(map[int64]*Node)
	o.rpcHandler = NewRpcHandler()
	o.hash = consistent.New()
	o.hash.NumberOfReplicas = 2
	o.hash.Add("1")
	o.store = s
	o.nextId = 2
	return o
}

func (o *Overlay) Start() *Overlay {
	o.rpcHandler.RegisterAction(Msg_JOIN, o.JoinAction)
	o.rpcHandler.RegisterAction(Msg_HEARTBEAT, o.HeartbeatAction)
	o.rpcHandler.RegisterAction(Msg_LEAVE, o.LeaveAction)
	o.rpcHandler.RegisterAction(Msg_GET, o.GetAction)
	o.rpcHandler.RegisterAction(Msg_PUT, o.PutAction)
	o.rpcHandler.RegisterInfoAction(Msg_NETWORK_INFO, o.NetworkInfoAction)
	go o.rpcHandler.receive(o.router)
	if o.me.addr != o.coordinator.addr {
		o.Join()
	} else {
		o.coordinator = o.me
		o.members[o.me.id] = o.me
		go o.checkAlive()
	}
	go o.debugMode()
	return o
}

func (o *Overlay) StartMember() *Overlay {
	log.Println("This Node is now a member")
	return o
}

func (o *Overlay) pinger(m *Node) {
	for _ = range time.Tick(3 * time.Second) {
		if _, ok := o.members[m.id]; !ok {
			break
		}
		o.Heartbeat(m)
	}
}

func (o *Overlay) checkAlive() {
	for _ = range time.Tick(time.Second) {
		for id, n := range o.members {
			if id != o.me.id && time.Since(o.members[n.id].lastAlive) > 10*time.Second {
				log.Printf("Member %d is now dead", n.id)
				delete(o.members, n.id)
				o.Leave(n)
				o.NetworkInfo()
			}
		}
	}
}

func (o *Overlay) debugMode() {
	for _ = range time.Tick(10 * time.Second) {
		log.Printf("Members: %s", o.members)
	}
}

func (o *Overlay) Join() {
	joinReq := new(JoinRequest)
	addr := o.me.addr.String()
	joinReq.Addr = &addr
	ch := o.Send(Msg_JOIN, joinReq, o.coordinator)
	go func() {
		respData := <-ch
		response := new(JoinResponse)
		proto.Unmarshal(respData, response)
		log.Println(response)
		o.coordinator = NodeFrom(response.Coordinator)
		switch response.GetStatus() {
		case JoinResponse_OK:
			if response.GetMembers() != nil {
				o.members = toMap(response.GetMembers())
			}
			o.me.id = response.GetYourId()
			o.StartMember()
		case JoinResponse_NOT_COORDINATOR:
			o.Join()
		}
	}()
}

func (o *Overlay) SetCoord() *Overlay {
	o.me.id = int64(1)
	o.coordinator = o.me
	return o
}

func (o *Overlay) JoinAction(reqData []byte) proto.Message {
	request := new(JoinRequest)
	proto.Unmarshal(reqData, request)
	response := new(JoinResponse)
	if o.coordinator.id == o.me.id {
		id := o.nextId
		o.nextId += 1
		addr := *request.Addr
		node := NodeFrom(&Member{Id: &id, UdpAddr: &addr})
		node.SetAlive()
		response.YourId = &node.id
		o.members[id] = node
		response.Status = JoinResponse_OK.Enum()
		response.Coordinator = o.coordinator.toMember()
		response.Members = toArray(o.members)
		o.hash.Add(strconv.Itoa(int(node.id)))
		o.NetworkInfo()
		go o.pinger(o.members[id])
	} else {
		response.Status = JoinResponse_NOT_COORDINATOR.Enum()
		response.Coordinator = o.coordinator.toMember()
	}
	return response
}

func (o *Overlay) Heartbeat(m *Node) {
	ch := o.Send(Msg_HEARTBEAT, new(HeartbeatRequest), m)
	go func() {
		<-ch
		o.members[m.id].SetAlive()
	}()
}

func (o *Overlay) HeartbeatAction(reqData []byte) proto.Message {
	request := new(HeartbeatRequest)
	proto.Unmarshal(reqData, request)
	return new(HeartbeatResponse)
}

func (o *Overlay) Leave(n *Node) {
	o.Send(Msg_LEAVE, &LeaveRequest{KickedMember: n.toMember(), Members: toArray(o.members)}, n)
}

func (o *Overlay) LeaveAction(reqData []byte) proto.Message {
	request := new(LeaveRequest)
	proto.Unmarshal(reqData, request)
	o.members = toMap(request.GetMembers())
	if request.GetKickedMember().GetId() == o.me.id {
		o.Join()
	}
	return new(LeaveResponse)
}

func (o *Overlay) Get(key string, subKey string) *GetResponse {
	node := o.GetNodeForKey(key)
	ch := o.Send(Msg_GET, &GetRequest{Key: &key, SubKey: &subKey}, node)
	respData := <-ch
	resp := new(GetResponse)
	proto.Unmarshal(respData, resp)
	return resp
}

func (o *Overlay) GetAction(data []byte) proto.Message {
	req := new(GetRequest)
	proto.Unmarshal(data, req)
	//if o.GetNodeForKey(req.GetKey()).id != o.me.id {
	//	return o.Get(req.GetKey(), req.GetSubKey())
	//}
	e, ok := o.store.Get(req.GetKey())
	if !ok {
		return &GetResponse{Status: GetResponse_NOT_FOUND.Enum()}
	}
	return &GetResponse{Status: GetResponse_OK.Enum(), Value: e.Value}
}

func (o *Overlay) Put(key string, subKey string, value []byte) *PutResponse {
	node := o.GetNodeForKey(key)
	ch := o.Send(Msg_PUT, &PutRequest{Key: &key, SubKey: &subKey, Value: value}, node)
	respData := <-ch
	resp := new(PutResponse)
	proto.Unmarshal(respData, resp)
	return resp
}

func (o *Overlay) PutAction(data []byte) proto.Message {
	req := new(PutRequest)
	proto.Unmarshal(data, req)
	//if o.GetNodeForKey(req.GetKey()).id != o.me.id {
	//	return o.Put(req.GetKey(), req.GetSubKey(), req.GetValue())
	//}
	o.store.Put(req.GetKey(), &store.Entry{Value: req.GetValue()})
	return new(PutResponse)
}

func (o *Overlay) NetworkInfo() {
	o.BroadcastInfo(Msg_NETWORK_INFO, &NetworkInfo{Coordinator: o.coordinator.toMember(), Members: toArray(o.members)})
}

func (o *Overlay) NetworkInfoAction(infoData []byte) {
	info := new(NetworkInfo)
	proto.Unmarshal(infoData, info)
	//o.members = toMap(info.GetMembers())
	log.Printf("NetworkInfo: %s", info)
	log.Printf("toMap: %s: %s", info.Members, o.members)
	o.members = toMap(info.GetMembers())
	o.coordinator = NodeFrom(info.GetCoordinator())
}

func (o *Overlay) Send(cmd Msg_Cmd, msg proto.Message, n *Node) chan []byte {
	return o.rpcHandler.Send(cmd, msg, n.addr, o.router)
}

func (o *Overlay) Broadcast(cmd Msg_Cmd, msg proto.Message) chan []byte {
	ch := make(chan []byte, len(o.members))
	for _, n := range o.members {
		go func() {
			b := <-o.rpcHandler.Send(cmd, msg, n.addr, o.router)
			ch <- b
		}()
	}
	return ch
}

func (o *Overlay) BroadcastInfo(cmd Msg_Cmd, msg proto.Message) {
	for _, n := range o.members {
		if n.id != o.me.id {
			o.rpcHandler.SendInfo(cmd, msg, n.addr, o.router)
		}
	}
}

func toMap(members []*Member) map[int64]*Node {
	result := make(map[int64]*Node)
	for _, m := range members {
		result[m.GetId()] = NodeFrom(m)
	}
	return result
}

func toArray(members map[int64]*Node) []*Member {
	result := make([]*Member, 0)
	for _, n := range members {
		result = append(result, n.toMember())
	}
	return result
}

func hash(s string) uint32 {
	h := crc32.NewIEEE()
	h.Write([]byte(s))
	return h.Sum32()
}

func (o *Overlay) GetNodeForKey(key string) *Node {
	s, err := o.hash.Get(key)
	if err != nil {
		log.Printf("No members in hash ring")
		return nil
	}
	id, err := strconv.Atoi(s)
	node, ok := o.members[int64(id)]
	log.Printf("Key %s gives node id %d", key, node.id)
	if !ok {
		log.Printf("No member with id: %d", id)
		return nil
	}
	return node
}
