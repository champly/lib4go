package encoding

import "testing"

func TestJSON2Struct(t *testing.T) {
	data := `{"a":"b"}`

	t.Log(JSON2Struct(data))
}

func TestYAML2Struct(t *testing.T) {
	data := `
name: "mosn-demo:20882"
virtual_hosts:
- name: "mosn.io.dubbo.DemoService:20882"
  retry_policy:
    num_retries: 3
  routes:
  - match:
      prefix: "/"
    route:
      timeout: 10s
      cluster: "outbound|20882||mosn.io.dubbo.DemoService"
      retry_policy:
        num_retries: 5
        per_try_timeout: 3s
`

	t.Log(YAML2Struct(data))
}
