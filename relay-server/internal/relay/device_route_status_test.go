package relay

// B-S2（方案 §B.3.1/§B.3.5，OD-3）定向用例：GET /v1/routes/{routeID}/devices/
// {deviceID}/status——设备鉴权的 route 状态查询。
//
// 覆盖：bridge set/remove 两态（bridgeOnline true/false，恰单字段响应＋no-store）、
// 设备鉴权负例（无/错 token → 401）、限流（超出 per-min 限额 → 429）。

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestDeviceRouteStatusReportsBridgePresence(t *testing.T) {
	_, ts := newTestServer(t, 600)
	credentials := provisionDevice(t, ts.URL)
	address := ts.URL + "/v1/routes/" + credentials.routeID + "/devices/phone-1/status"

	// 无 bridge：false。
	response, data := requestJSON(t, http.MethodGet, address, credentials.deviceAuth, nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status (no bridge) = %d body=%s", response.StatusCode, data)
	}
	if cc := response.Header.Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", cc)
	}
	var payload map[string]bool
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("payload not json: %v (%s)", err, data)
	}
	if len(payload) != 1 || !payload["bridgeOnline"] == false {
		t.Fatalf("payload = %v, want exactly {bridgeOnline:false}", payload)
	}

	// bridge 在线：true（bridge socket 挂起期间）。setBridge 在 bridge socket
	// 升级后异步生效——轮询至 true（上限 2s）。
	bridge := wsDial(t, ts.URL, "/v1/routes/"+credentials.routeID+"/bridge", credentials.bridgeAuth)
	defer bridge.Close()
	deadline := time.Now().Add(2 * time.Second)
	for {
		response, data = requestJSON(t, http.MethodGet, address, credentials.deviceAuth, nil)
		var online map[string]bool
		if response.StatusCode == http.StatusOK && json.Unmarshal(data, &online) == nil && online["bridgeOnline"] == true {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("bridgeOnline never became true: status=%d body=%s", response.StatusCode, data)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// bridge 断开：回到 false。
	bridge.Close()
	deadline = time.Now().Add(2 * time.Second)
	for {
		response, data = requestJSON(t, http.MethodGet, address, credentials.deviceAuth, nil)
		_ = response
		var gone map[string]bool
		if err := json.Unmarshal(data, &gone); err == nil && gone["bridgeOnline"] == false {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("bridgeOnline never returned to false: %s", data)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestDeviceRouteStatusRequiresDeviceAuth(t *testing.T) {
	_, ts := newTestServer(t, 600)
	credentials := provisionDevice(t, ts.URL)
	address := ts.URL + "/v1/routes/" + credentials.routeID + "/devices/phone-1/status"

	// 无 token → 401。
	response, data := requestJSON(t, http.MethodGet, address, "", nil)
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no-token status = %d body=%s", response.StatusCode, data)
	}
	// 错 token（bridgeAuth 不是设备凭据）→ 401。
	response, data = requestJSON(t, http.MethodGet, address, credentials.bridgeAuth, nil)
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong-token status = %d body=%s", response.StatusCode, data)
	}
	// 他设备 id → 401（deviceAuth 只对本 deviceID 有效）。
	response, data = requestJSON(t, http.MethodGet, ts.URL+"/v1/routes/"+credentials.routeID+"/devices/other-device/status", credentials.deviceAuth, nil)
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("cross-device status = %d body=%s", response.StatusCode, data)
	}
}

func TestDeviceRouteStatusRateLimited(t *testing.T) {
	_, ts := newTestServer(t, 3)
	credentials := provisionDevice(t, ts.URL)
	address := ts.URL + "/v1/routes/" + credentials.routeID + "/devices/phone-1/status"

	sawOK := false
	for i := 0; i < 12; i++ {
		response, data := requestJSON(t, http.MethodGet, address, credentials.deviceAuth, nil)
		switch response.StatusCode {
		case http.StatusOK:
			sawOK = true
		case http.StatusTooManyRequests:
			if !sawOK {
				t.Fatalf("rate limited before any 200 (i=%d) body=%s", i, data)
			}
			return // 限额触发即通过
		default:
			t.Fatalf("unexpected status %d body=%s", response.StatusCode, data)
		}
	}
	t.Fatal("12 requests never hit the rate limit (limiter family not applied?)")
}

// R1D-2（audit-bs2-r2）：不存在 routeID → 401——不泄露 route 存在性（与
// device socket 同口径：凭据对未知 route 一律按认证失败应答）。
func TestDeviceRouteStatusUnknownRouteIsUnauthorized(t *testing.T) {
	_, ts := newTestServer(t, 600)
	credentials := provisionDevice(t, ts.URL)
	response, data := requestJSON(t, http.MethodGet, ts.URL+"/v1/routes/rt_does_not_exist/devices/phone-1/status", credentials.deviceAuth, nil)
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unknown route status = %d body=%s, want 401（不泄露存在性）", response.StatusCode, data)
	}
}
