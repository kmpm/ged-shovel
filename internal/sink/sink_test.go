package sink

import "testing"

func Test_subjectify(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test1", args{"https://eddn.edcd.io/schemas/journal/1"}, "eddn.journal.1"},
		{"test2", args{"https://eddn.edcd.io/schemas/fssdiscoveryscan/1"}, "eddn.fssdiscoveryscan.1"},
		{"test3", args{"https://eddn.edcd.io/schemas/fsssignaldiscovered/1"}, "eddn.fsssignaldiscovered.1"},
		{"test4", args{"https://eddn.edcd.io/schemas/codexentry/1"}, "eddn.codexentry.1"},
		{"test5", args{"https://eddn.edcd.io/schemas/navroute/1"}, "eddn.navroute.1"},
		{"test6", args{"https://eddn.edcd.io/schemas/commodity/3"}, "eddn.commodity.3"},
		{"test7", args{"https://eddn.edcd.io/schemas/outfitting/2"}, "eddn.outfitting.2"},
		{"test8", args{"https://eddn.edcd.io/schemas/shipyard/2"}, "eddn.shipyard.2"},
		{"test9", args{"https://eddn.edcd.io/schemas/fssbodysignals/1"}, "eddn.fssbodysignals.1"},
		{"test10", args{"https://eddn.edcd.io/schemas/scanbarycentre/1"}, "eddn.scanbarycentre.1"},
		{"test11", args{"https://eddn.edcd.io/schemas/dockinggranted/1"}, "eddn.dockinggranted.1"},
		{"test12", args{"https://eddn.edcd.io/schemas/fssallbodiesfound/1"}, "eddn.fssallbodiesfound.1"},
		{"test13", args{"https://eddn.edcd.io/schemas/approachsettlement/1"}, "eddn.approachsettlement.1"},
		{"test14", args{"https://eddn.edcd.io/schemas/navbeaconscan/1"}, "eddn.navbeaconscan.1"},
		{"test15", args{"https://eddn.edcd.io/schemas/navbeaconscan/2/test"}, "eddn.navbeaconscan.2.test"},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := subjectify(tt.args.schema); got != tt.want {
				t.Errorf("subjectify() = %v, want %v", got, tt.want)
			}
		})
	}
}
