# kubevirt-sidecars
A repo for sample kubevirt sidecars
Remember this doesn't work on the Vanilla Redhat supplied Qemu (Qemu may not support the vmxnet3 interface)

## Build
Do 
```
go build .
```

and run
```
docker build . -t vmxnet3-hook
docker tag vmxnet3-hook <user-name>/vmxnet3-hook
docker push <user-name>/vmxnet3-hook
```

in the VMI definition just add an annotation
```
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  name: esx-1
  annotations:
    hooks.kubevirt.io/hookSidecars: '[{"image":"rparikh/vmxnet3-hook:latest"}]'
spec:
  domain:
    cpu:
      cores: 2
      model: host-model
    devices:
      disks:
        - disk:
            bus: sata
          name: volume-1
        - disk:
            bus: sata
          name: volume-2
      interfaces:
        - masquerade: {}
          name: default
    resources:
      requests:
        memory: 16Gi
  networks:
    - name: default
      pod: {}
  volumes:
    - name: volume-1
      persistentVolumeClaim:
        claimName: esxi-6.7
    - name: volume-2
      persistentVolumeClaim:
        claimName: esxihd
~
```