package timer

// func TestReserveResources(t *testing.T) {
// 	transformer := blueprints.Transformer{
// 		Name: "test",
// 		Cost: []blueprints.TransformerCost{
// 			{
// 				Resource:  string(common.Wood),
// 				Amount:    10,
// 				Temporary: false,
// 			},
// 		},
// 	}

// 	dataStruct := &structpb.Struct{
// 		Fields: map[string]*structpb.Value{
// 			"cost":   structpb.NewListValue(transformer.CostStructList()),
// 			"result": structpb.NewListValue(transformer.ResultStructList()),
// 		},
// 	}
// 	timer := Timer{
// 		Data:        dataStruct.AsMap(),
// 		InventoryID: uuid.Nil.String(),
// 	}
// 	g := &Grain{
// 		timer: &timer,
// 	}

// 	res := g.reserveResources()
// 	log.Printf("%#v", res)
// }
