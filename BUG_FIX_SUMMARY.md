# Bug Fix: HTML Rendering Duplication Issue

## Summary
Fixed a critical bug where HTML content was being rendered multiple times, causing text to appear duplicated on the page. The paragraph text was appearing ~15 times, headings appeared twice, and links appeared twice.

## Problem Analysis

### Symptoms
When rendering the following HTML:
```html
<div>
  <h1>Example Domain</h1>
  <p>This domain is for use in documentation examples without needing permission. Avoid use in operations.</p>
  <p><a href="https://iana.org/domains/example">Learn more</a></p>
</div>
```

**Expected**: 3 rendered elements (1 heading, 1 paragraph, 1 link)
**Actual**: 19 rendered elements with massive duplication

### Root Cause
The bug was in how the layout engine handled inline content that wrapped across multiple lines:

1. **Inline Layout Engine** (correct behavior):
   - When text wraps, creates one `InlineBox` per word/segment
   - Example: "Example Domain" → 2 InlineBoxes: ["Example", "Domain"]
   - Each InlineBox has the NodeID of its parent text node

2. **Layout Engine** (incorrect behavior):
   - For each InlineBox, created a separate child `LayoutBox`
   - All these LayoutBoxes had the same NodeID (the text node's ID)
   - Example: NodeID 4 would have 2 child LayoutBoxes, both with NodeID 4

3. **Display List Builder** (incorrect behavior):  
   - Iterated through all child LayoutBoxes
   - For each box, looked up the RenderNode by NodeID
   - Created a paint command with the FULL text from the RenderNode
   - Result: Each word segment caused the entire text to be rendered again

## Solution

### 1. Layout Engine Changes (`internal/renderer/layout.go`)
**Before**:
```go
// Create layout boxes for each inline box in the lines
for _, line := range lines {
    for _, inlineBox := range line.InlineBoxes {
        childLayoutBox := NewLayoutBox(inlineBox.NodeID)
        childLayoutBox.Display = DisplayInline
        childLayoutBox.Box = Rect{...}
        layoutBox.AddChild(childLayoutBox)  // ❌ Creates duplicates
        le.nodeMap[inlineBox.NodeID] = childLayoutBox
    }
}
```

**After**:
```go
// Store line boxes in the layout box
layoutBox.LineBoxes = lines

// DO NOT create child LayoutBox instances for inline boxes
// The LineBoxes contain all the information needed for rendering
// However, we still need to populate nodeMap for GetLayoutBox to work
processedNodeIDs := make(map[int64]bool)
for _, line := range lines {
    for _, inlineBox := range line.InlineBoxes {
        if !processedNodeIDs[inlineBox.NodeID] {
            processedNodeIDs[inlineBox.NodeID] = true
            // Map the inline node ID to the parent layout box
            le.nodeMap[inlineBox.NodeID] = layoutBox
        }
    }
}
```

### 2. Display List Builder Changes (`internal/renderer/display_list.go`)
**Before**:
```go
// Generate paint command based on node type
if renderNode.Type == NodeTypeText {
    dlb.addTextCommand(layoutBox, renderNode, displayList)
}

// Process children
for _, child := range layoutBox.Children {
    dlb.buildRecursive(child, renderMap, displayList)  // ❌ Processes duplicates
}
```

**After**:
```go
// Check if this layout box has inline content (LineBoxes)
if len(layoutBox.LineBoxes) > 0 {
    // Group inline boxes by NodeID to avoid duplicates
    processedNodes := make(map[int64]bool)
    
    for _, lineBox := range layoutBox.LineBoxes {
        for _, inlineBox := range lineBox.InlineBoxes {
            if processedNodes[inlineBox.NodeID] {
                continue  // ✅ Skip already processed nodes
            }
            processedNodes[inlineBox.NodeID] = true
            
            // Create paint command for the full text of the node
            cmd := &PaintCommand{
                Text: inlineRenderNode.Text,  // ✅ Full text, rendered once
                ...
            }
            displayList.AddCommand(cmd)
        }
    }
}
```

### 3. Test Updates
Updated tests that expected the old behavior (child LayoutBoxes for inline content):
- `TestComputeLayout`: Now checks for `LineBoxes` instead of `Children`
- `TestGetLayoutBox`: Accepts that inline nodes map to parent box
- `TestHitTest`: Uses block children instead of inline children
- `TestInlineLayoutIntegration`: Expects no child boxes for inline-only content
- `TestBlockWithInlineChildren`: Expects no child boxes in paragraphs

### 4. Regression Test
Added `TestBugFixDuplicateRendering` with the exact HTML from the bug report to ensure this issue doesn't recur.

## Impact

### Before Fix
- 19 rendered objects for simple HTML
- Massive text duplication
- Poor performance due to rendering overhead
- Confusing user experience

### After Fix
- 3 rendered objects (correct count)
- No duplication
- Improved performance
- Clean rendering matching expected output

## Testing
- All existing renderer tests pass
- New regression test added and passing
- CodeQL security scan: 0 vulnerabilities

## Files Changed
- `internal/renderer/layout.go` - Fixed duplicate LayoutBox creation
- `internal/renderer/display_list.go` - Fixed duplicate paint command generation
- `internal/renderer/layout_test.go` - Updated tests for new behavior
- `internal/renderer/inline_layout_integration_test.go` - Updated tests
- `internal/renderer/bug_regression_test.go` - Added regression test
