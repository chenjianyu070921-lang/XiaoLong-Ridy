package proto

import "testing"

func TestDriversvcExternalRPCMethodsAreUniqueAndRegistered(t *testing.T) {
	required := map[string]bool{
		"CreateDriver":        false,
		"RegisterDriver":      false,
		"UpdateDriver":        false,
		"GetDriver":           false,
		"GetDriverByPhone":    false,
		"SetDriverOnline":     false,
		"SetDriverOffline":    false,
		"ReportLocation":      false,
		"Heartbeat":           false,
		"CreateVehicle":       false,
		"UpdateVehicle":       false,
		"DeleteVehicle":       false,
		"GetVehicle":          false,
		"ListNearbyDrivers":   false,
		"GetDriverAiScore":    false,
		"UploadCertification": false,
		"GetCertification":    false,
		"CreateWithdraw":      false,
		"ListWithdraws":       false,
	}
	seen := make(map[string]struct{}, len(Driversvc_ServiceDesc.Methods))
	for _, method := range Driversvc_ServiceDesc.Methods {
		if _, ok := seen[method.MethodName]; ok {
			t.Fatalf("duplicate driversvc RPC method %q", method.MethodName)
		}
		seen[method.MethodName] = struct{}{}
		if _, ok := required[method.MethodName]; ok {
			required[method.MethodName] = true
		}
	}
	for method, found := range required {
		if !found {
			t.Fatalf("driversvc RPC method %q is not registered", method)
		}
	}
}
