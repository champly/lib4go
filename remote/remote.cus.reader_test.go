package remote

import (
	"fmt"
	"testing"
	"time"
)

var rsaPriv = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA3+ONOjcrnAeLayEvwZlDB8EKbylvLUas+iKu3R0PYps+WeY8
TGKvUmzFIPba8qBbHvw0sAeS81y/dhOVIO+IMaUNTRaBoQpfsxh6qTZIbdBWbyKI
zmub9zfOe+64uOkpv0Uq4LPypvZ9JGn0+UY9APb6+D9Mp6NR35FJLyNYZnZcz4Mu
zJ4wCGRgim1+1I9h1xXCPlDUrqo5+178vNfD0sp7exXLEK8r1E+0pkPeSbQTUuQ7
fDMZf3m5zSsYZyzaqASn8oLzPINa3F1A4spFk4+qKthu8/B/zxPAqkmphqHVDW0N
CeROuM4nmdEwegGXYK2SK4Iyxh8avOKki0T5kwIDAQABAoIBAG0GeH2C5D+lBOV0
UbcrFRMvlA8x1CvuIMnmHdUbE6TnCGPq1C42WD4BLbWxwEkqgXUDR/z4kzzjS3EK
dDKHsoDKaUHC1fk//f5Oy1yfTIH9VDnmTUyH5nlquahsRZP2Jxg3bHvj5SQdIC+d
UWgaJhbULr64xHFV/MasD0FfKuspizknDNSHfJz/oGeLNKCs7DTVNkgEN++Ec87J
KXnPUscrGmg64sL0HQpXB6PrGVpqEgwnXa8xruQwTBqjrwXfJIjoUIWCmhGUBzkP
CCO/h5W5HV9FnSHufJzeUblOFhD2KfRJ0tvu3u9oQlBsSRUQoa064AGAAK8z+uHm
/Ndz+NECgYEA8INRaHYXXrtAAs4hvO7nYHOiatP9nc/KA9VHgQDx2BjH/XjD5W+L
HkBcNhixFQPUPb1ejyy4c9zytLel7CKr9Z8lWEAl8hEoITMV4LrKF2MqW2aMKwkW
JxGN4Mox0qlAe4wimgCkX1FPEGaynn/9WJ8xy7Rg/DZUVxPG2y1Tr88CgYEA7k4y
l9UWljoIZEmrxipfXwSf1gzB8b4xYQnEkgavSg5hpbibgsoC/PxX2B03rubnWtYd
z65Y3AnBXcdocUGtt0NkCmRSEQjYyczHuhqPKXp2EF88AqhbdQ/EviPDttmwQiSD
6BayL/GJsoYOuT1FiOBIJ5ZdmRKiZi2kSp9wpv0CgYEAgUZ0SWvAIAER5PAbHkxj
PWqqEDWmCl8XvHu1FVgGphqb1FhHI1mTM01wwvr+o8cNG6pf2yE0e8J1CkH0AzqX
p0xFbGv+eWBTa5Tj24lK+ssohzxVdwRJTfKXig3kPdEPgdjO+GwD7d/sWWp588vj
xvC6eT2ZK7egGbXdmw1//+0CgYBVuS177rxUSAXyxYmUHHP4QzqYDjjKFEfBB3l9
qgfuVOQNcC4Iy1Bt3vxekowQT6GTzIgmyCnQ5XV4nZ3Vd/HchdJ75oCa/hq15QNH
z/wFyLalxwxYTGWx4307hLQHl6FO+cG1gEyS8Ik+/fhX7FGSHlP2YaHDya8/oFWE
PnyQpQKBgA4CawATyE62KypBggsw7liTManVXX8e6qiRE2nt8QHcsrzb6XXnoc42
IjZvBOGwpri3EEdLfREbHcZCRqbjxoiJeylfPZDr4eBEbUl7wL4xYscrPCm7m3X2
uEYY6WFwuNhLoOyaZ2b0cs1+W7JEKdpbsGoZrx384gKkp+RxOlaF
-----END RSA PRIVATE KEY-----`

func TestCusReader(t *testing.T) {
	client, err := NewCusReader(&ServerInfo{
		Host: "10.12.192.131",
		User: "root",
		Key:  rsaPriv,
		Port: 22,
	})
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("connect success")

	go func() {
		for {
			r := client.ExecGetResult()
			fmt.Println("go func:", r)
			time.Sleep(1 * time.Second)
		}
	}()

	// exec cmd
	// r, err := client.Exec("for i in `seq 1 10`;do echo $i; sleep 1;done")
	r, err := client.Exec("docker pull nginx")
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(r)
}
