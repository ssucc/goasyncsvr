package netinterface

type IConnector interface {
	WriteToChan([]byte) error
	ReadFromChan() ([]byte, error)
}

type ISvrSocket interface {
	WriteToChan([]byte) error
}

type ICliSocket interface {
	Connect() error
}
