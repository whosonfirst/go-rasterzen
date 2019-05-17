# go-whosonfirst-aws

There are many AWS wrappers. This one is ours.

## Install

You will need to have both `Go` (specifically [version 1.12](https://golang.org/dl/) or higher because we're using [Go modules](https://github.com/golang/go/wiki/Modules)) and the `make` programs installed on your computer. Assuming you do just type:

```
make tools
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Important

This works. Until it doesn't. It has not been properly documented yet.

## DSN strings

```
bucket=BUCKET region={REGION} prefix={PREFIX} credentials={CREDENTIALS}
```

Valid credentials strings are:

* `env:`

* `iam:`

* `{PATH}:{PROFILE}`

## See also

* https://docs.aws.amazon.com/sdk-for-go/

