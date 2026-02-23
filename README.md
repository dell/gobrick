# GOBRICK
**Library for iSCSI/FC/NVMe volume connection**


### Example

```
connector := gobrick.NewISCSIConnector(gobrick.ISCSIConnectorParams{})
dev, err := connector.ConnectVolume(context.Background(),
		gobrick.ISCSIVolumeInfo{
			Lun:1,
			Targets: []gobrick.ISCSITargetInfo{
				{Portal: "1.1.1.1",
				 Target: "iqn.2015-10.com.dell:dellemc-array-fnm00185000782-b-5dc4fceb"}}})

err = connector.DisconnectVolumeByDeviceName(context.Background(), "dm-1")
```
