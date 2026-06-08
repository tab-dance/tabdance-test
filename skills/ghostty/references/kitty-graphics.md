# Kitty Graphics — Image Protocol Support

In-depth reference for Kitty graphics protocol support in go-libghostty: image storage, placement iteration, and image/placement metadata.

## Overview

go-libghostty provides full support for the Kitty Graphics Protocol, allowing inspection of images stored in the terminal. This is useful for testing applications that render images in the terminal.

## KittyGraphics — Borrowed Image Storage

```go
type KittyGraphics struct { ptr C.GhosttyKittyGraphics }
```

Borrowed from the terminal via `term.KittyGraphics()`. Valid only until the next mutating terminal call.

### Obtaining

```go
func (t *Terminal) KittyGraphics() (*KittyGraphics, error)
```

### Placement Iterator

```go
func (kg *KittyGraphics) PlacementIterator(iter *KittyGraphicsPlacementIterator) error
```

Creates a placement iterator for iterating all image placements on the active screen.

## KittyGraphicsImage — Image Metadata

```go
type KittyGraphicsImage struct { /* internal */ }
```

### Methods

| Method | Return | Description |
|--------|--------|-------------|
| `ID()` | `uint32` | Image ID |
| `Number()` | `uint32` | Image number |
| `Width()` | `uint32` | Width in pixels |
| `Height()` | `uint32` | Height in pixels |
| `Format()` | `KittyGraphicsImageFormat` | Image format |
| `Compression()` | `KittyGraphicsImageCompression` | Compression method |
| `Data()` | `[]byte` | Raw image data (borrowed) |
| `Info(keys, values)` | `error` | Batch read |
| `GetMulti(keys, values)` | `error` | Batch read |

### Image Formats

| Format | Description |
|--------|-------------|
| `KittyGraphicsImageFormatRGB` | Raw RGB pixels |
| `KittyGraphicsImageFormatRGBA` | Raw RGBA pixels |
| `KittyGraphicsImageFormatPNG` | PNG encoded |
| `KittyGraphicsImageFormatGrayAlpha` | Grayscale + alpha |
| `KittyGraphicsImageFormatGray` | Grayscale |

### Compression Methods

| Method | Description |
|--------|-------------|
| `KittyGraphicsImageCompressionNone` | No compression |
| `KittyGraphicsImageCompressionZlib` | Zlib/Deflate compressed |

## KittyGraphicsPlacementIterator

```go
type KittyGraphicsPlacementIterator struct { /* internal */ }
```

### Methods

| Method | Return | Description |
|--------|--------|-------------|
| `SetLayer(layer)` | — | Filter by placement layer |
| `Next()` | `bool` | Advance to next placement (false = done) |
| `ImageID()` | `uint32` | Current placement's image ID |
| `PlacementID()` | `uint32` | Current placement ID |
| `IsVirtual()` | `bool` | Whether this is a virtual placement |
| `Image(term)` | `*KittyGraphicsImage` | Get the image for current placement |

### Placement Layers

| Layer | Description |
|-------|-------------|
| `KittyGraphicsPlacementLayerAll` | All layers |
| `KittyGraphicsPlacementLayerBelowBG` | Below background |
| `KittyGraphicsPlacementLayerBelowText` | Below text |
| `KittyGraphicsPlacementLayerAboveText` | Above text |

## Terminal Options for Kitty Graphics

| Method | Description |
|--------|-------------|
| `SetKittyImageStorageLimit(limit *uint64)` | Max image storage in bytes |
| `SetKittyImageMediumFile(enabled bool)` | Enable file medium |
| `SetKittyImageMediumTempFile(enabled bool)` | Enable temp file medium |
| `SetKittyImageMediumSharedMem(enabled bool)` | Enable shared memory medium |
| `SetAPCMaxBytes(limit *uint)` | APC buffer limit |
| `SetAPCMaxBytesKitty(limit *uint)` | Kitty APC buffer limit |

## Common Patterns

### Check for Images on Screen

```go
kg, _ := term.KittyGraphics()
var iter libghostty.KittyGraphicsPlacementIterator
kg.PlacementIterator(&iter)

for iter.Next() {
    imgID := iter.ImageID()
    placementID := iter.PlacementID()
    fmt.Printf("Image %d, Placement %d, Virtual: %v\n", imgID, placementID, iter.IsVirtual())
}
```

### Get Image Data

```go
kg, _ := term.KittyGraphics()
var iter libghostty.KittyGraphicsPlacementIterator
kg.PlacementIterator(&iter)

if iter.Next() {
    img := iter.Image(term)
    fmt.Printf("Image %d: %dx%d, format=%d\n", img.ID(), img.Width(), img.Height(), img.Format())
    data := img.Data()
    // Process raw image data
}
```
