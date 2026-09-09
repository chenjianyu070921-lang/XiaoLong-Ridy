package proto

import "testing"

func TestDriverServiceExternalRPCMethodsAreUniqueAndRegistered(t *testing.T) {
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
		"LoginBySms":          false,
		"UploadCertification": false,
		"GetCertification":    false,
		"CreateWithdraw":      false,
		"ListWithdraws":       false,
	}
	seen := make(map[string]struct{}, len(DriverService_ServiceDesc.Methods))
	for _, method := range DriverService_ServiceDesc.Methods {
		if _, ok := seen[method.MethodName]; ok {
			t.Fatalf("duplicate driver RPC method %q", method.MethodName)
		}
		seen[method.MethodName] = struct{}{}
		if _, ok := required[method.MethodName]; ok {
			required[method.MethodName] = true
		}
	}
	for method, found := range required {
		if !found {
			t.Fatalf("driver RPC method %q is not registered", method)
		}
	}
}
