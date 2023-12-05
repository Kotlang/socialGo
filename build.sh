rm -Rf generated
mkdir generated

mkdir generated/social
cd social-model
protoc --go_out=../generated/social --go_opt=paths=source_relative \
    --go-grpc_out=../generated/social --go-grpc_opt=paths=source_relative \
    *.proto
cd ..

mkdir generated/auth
cd auth-model
protoc --go_out=../generated/auth --go_opt=paths=source_relative \
    --go-grpc_out=../generated/auth --go-grpc_opt=paths=source_relative \
    *.proto
cd ..

mkdir generated/notification
cd notification-model
protoc --go_out=../generated/notification --go_opt=paths=source_relative \
    --go-grpc_out=../generated/notification --go-grpc_opt=paths=source_relative \
    *.proto
cd ..

go build -mod=mod -o build/ .
