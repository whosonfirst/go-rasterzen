# systemd

First clone the `rasterd` unit file:

```
cp rasterd.service.example rasterd.service
```

Adjust the file to taste. It assumes as few things that you may need or want to tweak:

* That you've built and copied go-rasterzen/bin/rasterd to `/usr/local/bin/rasterd`
* That you want to run `rasterd` as user `www-data` (you will need to change this if you're running CentOS for example)
* That you will replace the `-stuff -stuff -stuff` flags in the `ExecStart` with meaningful config flags

Move the file in to place:

```
mv rasterd.service /lib/systemd/system/rasterd.service
```

Now tell `systemd` about it:

```
systemctl enable rasterd.service
sudo systemctl start rasterd
```

## See also

* https://fabianlee.org/2017/05/21/golang-running-a-go-binary-as-a-systemd-service-on-ubuntu-16-04/
