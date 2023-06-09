//go:build integration
// +build integration

package pkg

import (
	"fmt"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	connectorv1alpha "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

func TestTemp(t *testing.T) {
	config := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"api_token": {Kind: &structpb.Value_StringValue{StringValue: "<valid api key>"}},
			"task":      {Kind: &structpb.Value_StringValue{StringValue: "Text to Image"}},
			"engine":    {Kind: &structpb.Value_StringValue{StringValue: "stable-diffusion-v1-5"}},
		},
	}
	in := []*connectorv1alpha.DataPayload{{
		Texts: []string{"dog", "black"},
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{},
		},
	}}
	c := Init(nil, ConnectorOptions{})
	con, err := c.CreateConnection(c.ListConnectorDefinitionUids()[0], config, nil)
	fmt.Printf("err:%s", err)
	op, err := con.Execute(in)
	fmt.Printf("\n op :%v, err:%s", op, err)
}
