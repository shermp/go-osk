# go-osk
go-osk is an experimental on-screen keyboard, primarily aimed at those developing for Kobo ereaders.

This was dreamed up on a lazy Sunday afternoon. It may or may not evolve beyond its current form...

## Installation & Usage

go-osk can be obtained using go get:
```
go get github.com/shermp/go-osk/...
```

Refer to `example/main.go` for a basic usage example.

## Keymap JSON format
Keymaps are stored as a JSON file. The JSON gets converted to the osk at run-time. A sample keymap is included in the repository.

An abridged keymap is as follows:
```
{
	"lang": "en_us",
	"kbMargins": {
		"top": 0.60,
		"bottom": 0.05,
		"left": 0.0,
		"right": 0.0
	},
	"totalKeyWidth": 3,
	"totalRowHeight": 1.8,
	"rows": [
		{
			"rowHeight": 0.8,
			"keys": [
				{
					"isPadding": false,
					"keyType": 0,
					"keyWidth": 1,
					"char": "1"
				},
				{
					"isPadding": false,
					"keyType": 0,
					"keyWidth": 1,
					"char": "2"
				},
				{
					"isPadding": false,
					"keyType": 0,
					"keyWidth": 1,
					"char": "3"
				}
            ]
        },
        {
			"rowHeight": 1,
			"keys": [
				{
					"isPadding": true,
					"keyType": 0,
					"keyWidth": 1,
					"char": ""
				},
				{
					"isPadding": false,
					"keyType": 0,
					"keyWidth": 1,
					"char": "q"
				},
				{
					"isPadding": false,
					"keyType": 0,
					"keyWidth": 1,
					"char": "w"
				}
            ]
        }
    ]
}
```
The fields are as follows:

|Field|Value|
|---|---|
|"lang"|Keyboard language. Currently unused|
|"kbMargins"|`"top"` `"bottom"` `"left"` `"right"` margins. Each margin is expressed as a percentage of the screen dimensions. This controls the size and position of the onscreen keyboard. Valid values for each key are 0.0 - 0.8. Each top/bottom and left/right pair must sum to <= 0.8|
|"totalKeyWidth"|This is a somewhat arbitrary figure. If one assumes that a standard key is one unit wide, then this is how many units make up a row.|
|"totalRowHeight"|The same as above, but for rows.
|"rows"|An array of rows of the keyboard. Each row contains the following fields.|
|"rowHeight"|The height of the row. This is in proportion to the "totalRowHeight", where the sum of all "rowHeight" values must equal "totalRowHeight"|
|"keys"|Each row contains an array of keys, the fields for which are described below.|
|"isPadding"|This is used to tell the OSK that this 'key' is used merely for visual padding. "keyWidth" is the only other field used when this field is true.|
|"keyType"|This is the type of key, whether it be a standard character, or a control key such as "shift" "backspace" etc. Standard keys have the value 0.|
|"keyWidth"|This is the width of the key, and is proportional to "totalKeyWidth". All keys in a row must sum to "totalKeyWidth"|
|"char"|This is the printable character of the key.|