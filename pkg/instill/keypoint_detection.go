package instill

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

func (c *Connection) executeKeyPointDetection(model *modelPB.Model, inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, model.Name)
	}
	dataPayload := inputs[0]
	if len(dataPayload.Images) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", *dataPayload, model.Name)
	}
	base64Str, err := fetchImageFromURL(dataPayload.Images[0])
	if err != nil {
		return nil, fmt.Errorf("invalid image string: %v for model: %s", dataPayload.Images[0], model.Name)
	}
	modelInput := &modelPB.TaskInput_Keypoint{
		Keypoint: &modelPB.KeypointInput{Type: &modelPB.KeypointInput_ImageBase64{ImageBase64: base64Str}},
	}
	req := modelPB.TriggerModelRequest{
		Name:       model.Name,
		TaskInputs: []*modelPB.TaskInput{{Input: modelInput}},
	}
	if c.client == nil || c.client.GRPCClient == nil {
		return nil, fmt.Errorf("client not setup: %v", c.client)
	}
	res, err := c.client.GRPCClient.TriggerModel(context.Background(), &req)
	if err != nil || res == nil {
		return nil, err
	}
	output := res.GetTaskOutputs()
	if len(output) <= 0 {
		return nil, fmt.Errorf("invalid output: %v for model: %s", output, model.Name)
	}
	keyPointOutput := output[0].GetKeypoint()
	if keyPointOutput == nil {
		return nil, fmt.Errorf("invalid output: %v for model: %s", keyPointOutput, model.Name)
	}
	values := make([]*structpb.Value, 0, len(keyPointOutput.Objects))
	for _, o := range keyPointOutput.Objects {
		keyPoints := make([]*structpb.Value, 0, len(o.Keypoints))
		for _, k := range o.Keypoints {
			kp := &structpb.Value{
				Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"v": {Kind: &structpb.Value_NumberValue{NumberValue: float64(k.V)}},
							"x": {Kind: &structpb.Value_NumberValue{NumberValue: float64(k.X)}},
							"y": {Kind: &structpb.Value_NumberValue{NumberValue: float64(k.Y)}},
						},
					},
				},
			}
			keyPoints = append(keyPoints, kp)
		}
		obj := &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"score":     {Kind: &structpb.Value_NumberValue{NumberValue: float64(o.Score)}},
						"keypoints": {Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: keyPoints}}},
						"bounding_box": {Kind: &structpb.Value_StructValue{
							StructValue: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"top":    {Kind: &structpb.Value_NumberValue{NumberValue: float64(o.BoundingBox.Top)}},
									"left":   {Kind: &structpb.Value_NumberValue{NumberValue: float64(o.BoundingBox.Left)}},
									"width":  {Kind: &structpb.Value_NumberValue{NumberValue: float64(o.BoundingBox.Width)}},
									"height": {Kind: &structpb.Value_NumberValue{NumberValue: float64(o.BoundingBox.Height)}},
								},
							},
						},
						},
					},
				},
			},
		}
		values = append(values, obj)
	}
	inputs[0].StructuredData = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"objects": {Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: values}}},
		},
	}
	return inputs, nil
}