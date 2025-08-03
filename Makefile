# 生成gRPC相关文件
.PHONY: grpc
grpc:
	@buf format -w api/proto
	@buf lint api/proto
	@buf generate api/proto