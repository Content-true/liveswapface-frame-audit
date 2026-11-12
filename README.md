# Frame Audit for Live Face Swap Reviews

Frame Audit is a small, local Go command that compares consecutive PNG or JPEG frames and produces a CSV of normalized visual changes. It helps reviewers locate abrupt transitions in a recorded face-swap test. A high score can also come from ordinary movement or lighting changes, so every flagged frame needs human review.

## Use

```sh
go run . -dir ./frames -rect 240,80,400,400 -threshold 0.25 -out review.csv
```

Frame names are processed in lexicographic order; use zero-padded names such as `frame-0001.png`. Omit `-rect` to compare the whole frame. The CSV contains each frame pair, a score from 0 to 1, and a Boolean flag. The tool reads local images only and sends nothing to a server.

For a repeatable review, capture both an unprocessed baseline and the output seen by the receiving application. Use a consistent camera, crop, lighting setup and frame rate. The command can help narrow the moments worth inspecting when evaluating [live swap face](https://liveswapface.com/) footage, but it neither performs face swapping nor judges whether an image is authentic. Use only images you have permission to process, and label altered footage when sharing it.

## Requirements

Go 1.22 or newer. No third-party Go dependencies. The frames must have matching dimensions. `-rect` uses pixel coordinates and must fit within every frame.

## License

MIT; see [LICENSE](LICENSE).
