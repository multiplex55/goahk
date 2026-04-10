# UIA selectors with inspect-driven workflow

This flow combines `goahk-inspect` and the callback `ctx.UIA` service:

1. Discover stable properties for a target element:

```bash
goahk-inspect uia under-cursor
goahk-inspect uia under-cursor --copy-best-selector
goahk-inspect uia tree --active-window --depth 4
```

2. Review `SupportedPatterns` and ranked `SelectorSuggestions` (best to worst with rationale).
3. Optionally use `--copy-best-selector` to print `BestSelectorJSON` as a copy-ready line.
4. Capture `automationId`, `name`, and nearby ancestors.
5. Use those values in a callback selector.

```go
sel := goahk.SelectByAutomationID("submitBtn").WithAncestors(
    goahk.SelectByName("Checkout").WithControlType("Window"),
)
_, err := ctx.Automation.Invoke(sel, 3*time.Second, 100*time.Millisecond)
```

Practical tip: prefer `automationId` first, then add `Ancestors` for extra stability when multiple controls share names.
