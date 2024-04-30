package cmcd

import (
	"fmt"
	"net/url"
)

func ExampleParseInfo() {
	ex := "http://test.example.com/?CMCD=br%3D3200%2Cbs%2Cd%3D4004%2Cmtp%3D25400"
	u, err := url.Parse(ex)
	param := u.Query().Get("CMCD")
	fmt.Println(param)
	info, err := ParseInfo(param)
	if err != nil {
		// handle...
	}
	fmt.Println(info.Bitrate)
	// Output:
	// br=3200,bs,d=4004,mtp=25400
	// 3200
}
