// coding.go
package main

import "github.com/otsimo/gorsa"

type Encoder struct{ encoder gorsa.Encoder }
type Decoder struct{ decoder gorsa.Decoder }

func NewEncoder(data []byte, packetSize int) (*Encoder, error) {
	enc, err := gorsa.NewEncoder(data, uint16(packetSize))
	if err != nil { return nil, err }
	return &Encoder{encoder: enc}, nil
}

func (e *Encoder) GetEncodedPacket() []byte {
	return e.encoder.Encode()
}

func NewDecoder() *Decoder {
	return &Decoder{decoder: gorsa.NewDecoder()}
}

func (d *Decoder) AddPacket(packet []byte) bool {
	return d.decoder.Decode(packet)
}

func (d *Decoder) GetDecodedData() ([]byte, error) {
	data, ok := d.decoder.GetResult()
	if !ok { return nil, fmt.Errorf("decoding failed") }
	return data, nil
}
