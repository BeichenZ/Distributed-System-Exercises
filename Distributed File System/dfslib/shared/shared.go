package shared
//Store Shared Data Structure,interface between client and server
import "errors"

type Chunk [32]byte
type OneStringMsg struct{
	Msg string
}

type ExistsMsg struct {
	Exists bool
}
type GenericArgs struct {
	StringOne string
	StringTwo string
	BoolOne bool
	IntOne int
	IntArray []int
}
type GenericReply struct {
	BoolOne bool
	ChunkArray []Chunk
}
type ChunckUploadReply struct {
		
}
type Args struct {
	A,B int
}

type Quotient struct {
	Quo,Rem int
}

type Arith interface {
	Multiply(args *Args,reply *int) error
	Divide(args *Args,quo *Quotient) error
}
//DFSService Interface and Relevant Structs
type RNCArgs struct {
	LocalIP string
	LocalPath string
}
type RNCReply struct {
	ID int
}
type DFSService interface {
	RegisterNewClient(args *RNCArgs,reply *RNCReply) error
	GlobalFileExists(args *OneStringMsg, reply *ExistsMsg) error 
	UpdateFileInfo(args *GenericArgs, reply *GenericReply) error
}
type ClientService interface{
	UploadFollowingChunks(args *GenericArgs,reply *GenericReply) error
}
//One implementation of Arith Interface
type ArithObjT1 int
func (t *ArithObjT1) Multiply(args *Args,reply *int) error {
	*reply = args.A * args.B
	return nil
}
func (t *ArithObjT1) Divide(args *Args, quo *Quotient) error {

	if args.B == 0 {

		return errors.New("divide by zero")

	}

	quo.Quo = args.A / args.B

	quo.Rem = args.A % args.B

	return nil
}
