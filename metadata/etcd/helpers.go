package etcd

import (
	"bytes"
	"encoding/binary"
	"path"

	pb "github.com/barakmich/agro/internal/etcdproto/etcdserverpb"
	"github.com/barakmich/agro/models"
)

func mkKey(s ...string) []byte {
	s = append([]string{keyPrefix}, s...)
	return []byte(path.Join(s...))
}

func uint64ToBytes(x uint64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, x)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func bytesToUint64(b []byte) uint64 {
	r := bytes.NewReader(b)
	var out uint64
	err := binary.Read(r, binary.LittleEndian, &out)
	if err != nil {
		panic(err)
	}
	return out
}

type transact struct {
	tx *pb.TxnRequest
}

func tx() *transact {
	t := transact{&pb.TxnRequest{}}
	return &t
}

func (t *transact) If(comps ...*pb.Compare) *transact {
	t.tx.Compare = comps
	return t
}

func requestUnion(comps ...interface{}) []*pb.RequestUnion {
	var out []*pb.RequestUnion
	for _, v := range comps {
		switch m := v.(type) {
		case *pb.RangeRequest:
			out = append(out, &pb.RequestUnion{&pb.RequestUnion_RequestRange{RequestRange: m}})
		case *pb.PutRequest:
			out = append(out, &pb.RequestUnion{&pb.RequestUnion_RequestPut{RequestPut: m}})
		case *pb.DeleteRangeRequest:
			out = append(out, &pb.RequestUnion{&pb.RequestUnion_RequestDeleteRange{RequestDeleteRange: m}})
		default:
			panic("cannot create this request option within a requestUnion")
		}
	}
	return out
}

func (t *transact) Then(comps ...interface{}) *transact {
	ru := requestUnion(comps...)
	t.tx.Success = ru
	return t
}

func (t *transact) Else(comps ...interface{}) *transact {
	ru := requestUnion(comps...)
	t.tx.Failure = ru
	return t
}

func (t *transact) Tx() *pb.TxnRequest {
	return t.tx
}

func keyEquals(key []byte, value []byte) *pb.Compare {
	return &pb.Compare{
		Target: pb.Compare_VALUE,
		Key:    key,
		Result: pb.Compare_EQUAL,
		TargetUnion: &pb.Compare_Value{
			Value: value,
		},
	}
}

func keyExists(key []byte) *pb.Compare {
	return &pb.Compare{
		Target: pb.Compare_VERSION,
		Result: pb.Compare_GREATER,
		Key:    key,
		TargetUnion: &pb.Compare_Version{
			Version: 0,
		},
	}
}

func keyNotExists(key []byte) *pb.Compare {
	return &pb.Compare{
		Target: pb.Compare_VERSION,
		Result: pb.Compare_LESS,
		Key:    key,
		TargetUnion: &pb.Compare_Version{
			Version: 1,
		},
	}
}

func keyIsVersion(key []byte, version int64) *pb.Compare {
	return &pb.Compare{
		Target: pb.Compare_VERSION,
		Result: pb.Compare_EQUAL,
		Key:    key,
		TargetUnion: &pb.Compare_Version{
			Version: 1,
		},
	}
}

func setKey(key []byte, value []byte) *pb.PutRequest {
	return &pb.PutRequest{
		Key:   key,
		Value: value,
	}
}

func getKey(key []byte) *pb.RangeRequest {
	return &pb.RangeRequest{
		Key: key,
	}
}

func getPrefix(key []byte) *pb.RangeRequest {
	end := make([]byte, len(key))
	copy(end, key)
	end[len(end)-1]++
	return &pb.RangeRequest{
		Key:      key,
		RangeEnd: end,
	}
}

// *********

func newDirProto(md *models.Metadata) []byte {
	a := models.Directory{
		Metadata: md,
		Files:    make(map[string]uint64),
	}
	b, err := a.Marshal()
	if err != nil {
		panic(err)
	}
	return b
}