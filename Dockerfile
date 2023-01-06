FROM fedora:28

COPY kubevirt-sidecars /kubevirt-sidecars

ENTRYPOINT [ "/kubevirt-sidecars" ]
