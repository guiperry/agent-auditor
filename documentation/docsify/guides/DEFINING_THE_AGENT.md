### 🧠 Defining Features of an AI Agent
To be considered an AI Agent, a system typically includes:

1.  **Perception**  
    Ability to sense or receive input from the environment (e.g. via sensors, APIs, databases, files, UIs).
    
    *In code, this could mean:*
    - Reading from a camera, microphone, or log.
    - Ingesting user input or system state.

2.  **Reasoning / Decision-making**  
    Based on inputs, it can make decisions using:
    - Rule-based logic
    - Statistical models
    - Machine learning
    - Planning algorithms
    - Reinforcement learning

3.  **Memory or State**  
    Maintains internal state, history, or memory over time.
    
    *This allows it to:*
    - Adapt to past experiences
    - Handle tasks with temporal dependencies

4.  **Action / Output**  
    Takes actions that affect the environment, such as:
    - Sending messages
    - Updating files or databases
    - Calling APIs
    - Controlling devices

5.  **Autonomy**  
    Operates without constant human supervision.
    
    *May run continuously, periodically, or reactively.*

6.  **Learning (optional but common)**  
    Improves performance over time using machine learning or feedback mechanisms.
    
    *Not all AI agents learn, but many modern ones do.*

### 🧩 For a WASM File to Be an AI Agent
WASM (WebAssembly) is a binary format that enables code to run in a safe, fast, and portable way, often in browsers or sandboxed environments. A WASM file can contain or serve as a runtime for an AI agent if it includes the following functionality:

#### ✅ Core Capabilities in WASM
To qualify as an AI agent in WASM, the module must have:

| Functionality | Description |
| :--- | :--- |
| **Input interface** | Functions or APIs to receive data from the host environment |
| **Processing logic** | Encapsulated AI logic (rules, ML models, planners, etc.) |
| **State management** | Persistent or mutable memory across invocations |
| **Output interface**| Functions to return decisions, actions, or updates to the environment |
| **Autonomy logic**| Either internally driven (e.g., time/event triggers) or externally managed control flow that gives it independence |

> 🚫 **Not Enough:** A WASM file is **not** an AI agent just because:
> - It runs a neural network.
> - It does matrix operations.
> - It compresses or decompresses data.
> - It acts as a model inferencer (unless embedded in an autonomous loop).

#### ✅ Minimal Example: WASM AI Agent
Say a WASM module receives sensor data, decides whether to activate a cooling fan, and remembers past decisions. If it exposes:
- `sense()` to receive data
- `decide()` to apply logic
- `act()` to trigger outcomes
- And stores internal memory

...that qualifies as an AI agent, even if basic.

### Summary
| Criteria | AI Agent | WASM Requirements |
| :--- | :--- | :--- |
| **Perception** | Yes | Import functions to receive input |
| **Reasoning** | Yes | Internal logic (rules, ML, etc.) |
| **Action** | Yes | Export functions to act |
| **State** | Yes | Memory or mutable global variables |
| **Autonomy** | Yes | Triggered via environment or loop |
| **Learning** | Optional | Optional |

---

### 🧠 Example: Go WASM AI Agent (pre-compiled source):
```go
// This example demonstrates how to create a simple AI agent in Go/WASM.
//
// The agent has three primary functionalities:
//   1. Sense Temperature: Receives temperature readings from the environment.
//   2. Decide: Determines whether to turn on a cooling fan based on the current temperature.
//   3. Set Threshold: Allows changing the temperature threshold dynamically.
// The agent maintains its own state, including the current temperature and the threshold value.
// Usage:
//   - Call 'setThreshold' function to change the temperature threshold.
//   - Call 'senseTemperature' function to update the current temperature.
//   - Call 'decide' function to determine whether to turn on the cooling fan.
// Note: This example uses JavaScript's WebAssembly API to expose these functions to the browser.
//go:generate go build -o ai-agent.wasm -target=wasm -ldflags="-w -s" main.go
//go:build js && wasm

package main

import (
    "syscall/js"
    "sync"
)

var (
    mu        sync.Mutex
    threshold float64 = 25.0  // default threshold
    lastTemp  float64 = 0.0
)

// setThreshold sets the temperature threshold for the agent
func setThreshold(this js.Value, args []js.Value) interface{} {
    mu.Lock()
    defer mu.Unlock()
    if len(args) > 0 {
        threshold = args[0].Float()
    }
    return nil
}

// senseTemperature receives a temperature reading from the environment
func senseTemperature(this js.Value, args []js.Value) interface{} {
    mu.Lock()
    defer mu.Unlock()
    if len(args) > 0 {
        lastTemp = args[0].Float()
    }
    return nil
}

// decide returns true if the fan should be turned on
func decide(this js.Value, args []js.Value) interface{} {
    mu.Lock()
    defer mu.Unlock()
    shouldTurnOn := lastTemp > threshold
    return js.ValueOf(shouldTurnOn)
}

// getState returns internal memory/state (threshold and lastTemp)
func getState(this js.Value, args []js.Value) interface{} {
    mu.Lock()
    defer mu.Unlock()
    state := map[string]interface{}{
        "threshold": threshold,
        "lastTemp":  lastTemp,
    }
    return js.ValueOf(js.Global().Get("Object").New(state))
}

func registerCallbacks() {
    js.Global().Set("setThreshold", js.FuncOf(setThreshold))
    js.Global().Set("senseTemperature", js.FuncOf(senseTemperature))
    js.Global().Set("decide", js.FuncOf(decide))
    js.Global().Set("getState", js.FuncOf(getState))
}

func main() {
    c := make(chan struct{}, 0)
    registerCallbacks()
    <-c // keep WASM running
}
```

---

<div class="footer-links">
<a href="#/legal/CODE_OF_CONDUCT.md" class="footer-link">Contributor Covenant Code of Conduct</a> | <a href="#/legal/PRIVACY_POLICY.md" class="footer-link">PRIVACY_POLICY.md</a> | <a href="#/legal/TERMS_AND_CONDITIONS.md" class="footer-link">TERMS AND CONDITIONS</a>

© 2025 Agent Auditor
</div>
