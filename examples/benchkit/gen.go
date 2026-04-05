package benchkit

//go:generate go run ../../cmd/genjson -dir . -out zz_generated.genjson.go -types A,B,C,D,E,F,G,GItem,H,HTree -features unknown_fields,required_fields -emit-marshaler
