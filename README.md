# :lock: **Important Notice**
Starting with Container Storage Modules `v1.16.0`, this repository will transition to a closed-source model.<br>
* The current version remains open source and will continue to be available under the existing license.
* Customers will continue to receive access to enhanced features, timely updates, and official support through our commercial offerings.
* We remain committed to the open-source community - users engaging through Dell community channels will continue to receive guidance and support via Dell Support.

We sincerely appreciate the support and contributions from the community over the years.<br>
For access requests or inquiries, please contact the maintainers directly at [Dell Support](https://www.dell.com/support/kbdoc/en-in/000188046/container-storage-interface-csi-drivers-and-container-storage-modules-csm-how-to-get-support).

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
