package tests

var rawNodes = []*Node{
	{
		Name: "dev department",
		Children: []*Node{
			{
				Name: "dev center",
				Children: []*Node{
					{
						Name: "dev group 1",
						Children: []*Node{
							{
								Name: "dev team 1",
							},
							{
								Name: "dev team 2",
							},
						},
					},
					{
						Name: "dev group 2",
						Children: []*Node{
							{
								Name: "dev team 3",
							},
							{
								Name: "dev team 4",
							},
						},
					},
				},
			},
			{
				Name: "test center",
				Children: []*Node{
					{
						Name: "test group 1",
						Children: []*Node{
							{
								Name: "test team 1",
							},
							{
								Name: "test team 2",
							},
						},
					},
					{
						Name: "test group 2",
						Children: []*Node{
							{
								Name: "test team 3",
							},
							{
								Name: "test team 4",
							},
						},
					},
				},
			},
		},
	},
	{
		Name: "product department",
		Children: []*Node{
			{
				Name: "product center",
				Children: []*Node{
					{
						Name: "product group 1",
						Children: []*Node{
							{
								Name: "product team 1",
							},
							{
								Name: "product team 2",
							},
						},
					},
					{
						Name: "product group 2",
						Children: []*Node{
							{
								Name: "product team 3",
							},
							{
								Name: "product team 4",
							},
						},
					},
				},
			},
			{
				Name: "design center",
				Children: []*Node{
					{
						Name: "design group 1",
						Children: []*Node{
							{
								Name: "design team 1",
							},
							{
								Name: "design team 2",
							},
						},
					},
					{
						Name: "design group 2",
						Children: []*Node{
							{
								Name: "design team 3",
							},
							{
								Name: "design team 4",
							},
						},
					},
				},
			},
		},
	},
}
