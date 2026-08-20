module example.com/corp/gateway

go 1.22

require example.com/corp/auth v0.0.0

replace example.com/corp/auth => ../module_auth
