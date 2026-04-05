package basic

import "fmt"

//go:generate go run ../../cmd/genjson -dir . -out zz_generated.genjson.go -types User,UserPlain -features unknown_fields,required_fields -emit-marshaler

func Demo() error {
	body := []byte(`{"id":1,"name":"Artem","email":"a@example.com","nick":"miXeD"}`)

	u, err := DecodeUser(body)
	if err != nil {
		return err
	}
	fmt.Printf("decoded: %+v\n", u)

	encoded, err := EncodeUser(u)
	if err != nil {
		return err
	}
	fmt.Printf("encoded: %s\n", encoded)
	return nil
}
