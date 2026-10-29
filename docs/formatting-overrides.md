# Explicit run formatting

Bold, italic, strike, small caps, and all caps distinguish three states: omitted (inherit), true (on), and false (off). JSON preserves an explicit false through document serialization. The DOCX reader retains explicit off declarations, and the writer emits the corresponding property with `w:val="0"`.

Existing Go literals that set a toggle to true keep their behavior. A zero-valued boolean in a Go literal means inherit. Use `SetBold(false)`, `SetItalic(false)`, `SetStrike(false)`, `SetSmallCaps(false)`, or `SetAllCaps(false)` to request an explicit off value. The corresponding `Specified` method distinguishes off from inherit. Copying attributes preserves that distinction.

`RunAttrs.Overlay` applies direct run overrides to inherited attributes. Named paragraph styles inherit document run defaults and supported parent run attributes; an explicit off declaration overrides an inherited on declaration. Cyclic parent references fail resolution instead of recursing indefinitely.

Underline, baseline position, and character spacing also distinguish inheritance from an explicit zero. Use `SetUnderline(UnderlineNone)`, `SetBaseline(BaselineNone)`, and `SetTracking(0)` to remove an inherited underline, return to the normal baseline, or reset character spacing. Their corresponding `Specified` methods retain the distinction through JSON and DOCX. Use the corresponding `Inherit` method to remove an individual direct override and resume inheritance without changing other properties.

All eighteen standard DOCX underline values are preserved. The first five underline enum values keep their existing numeric identities. Additional styles are appended. Unknown styles and underline properties outside the model prevent a canonical save instead of silently becoming single underlines.

DOCX fidelity reports expose these capabilities:

| Capability | Preserved behavior |
| --- | --- |
| `explicit-run-toggles` | Explicit on/off declarations for the five boolean run properties |
| `explicit-run-values` | Explicit resets for underline, baseline, and character spacing |
| `underline-variants` | All eighteen standard underline values |

Consumers must check these capabilities before saving edits that older engines cannot preserve. An editing adapter must also support the requested style in its own renderer and conversion layer.

These capabilities cover the listed run declarations. Complete character-style toggle resolution, paragraph property inheritance, character styles applied over paragraph formatting, explicit black color overrides, and zero-valued paragraph spacing require their own presence and resolution coverage.

References:

- https://learn.microsoft.com/en-us/dotnet/api/documentformat.openxml.wordprocessing.bold?view=openxml-3.0.1
- https://learn.microsoft.com/en-us/openspecs/office_standards/ms-oi29500/bd62641d-7786-473f-872d-4b3205e91a82

- https://learn.microsoft.com/en-us/dotnet/api/documentformat.openxml.wordprocessing.underlinevalues?view=openxml-3.0.1
