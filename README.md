# azure-blobstore-resource
A concourse resource to interact with the azure blob service. Currently only supports versioning
blobs using snapshots.

> NOTE: The resource has been moved from the `czero` dockerhub account to the `pcfabr` dockerhub
> account. If your pipeline is currently using the resource from `czero` it should be switched
> to `pcfabr`.

## Source Configuration

* `storage_account_name`: *Required.* The storage account name on Azure.

* `storage_account_key`: *Required.* The storage account access key for the storage account on Azure.

* `container`: *Required.* The name of the container in the storage account.

* `versioned_file`: *Required.* The file name of the blob to be managed by the resource.
  The resource only pulls the latest snapshot. If the blob doesn't have a snapshot, the
  resource will not find the blob. A new snapshot must also be created when a blob is
  updated for the resource to successfully check new versions.

## Behavior

### `check`: Extract snapshot versions from the container.

Checks for new snapshot versions for a file. If a blob exists without a snapshot
the file will not be found.

### `in`: Fetch a blob from the container.

Places the blob file in the destination.

### `out`: Upload a blob to the container.

Uploads a file to the container. After uploading the blob it will create a
new snapshot of the blob.

#### Parameters

* `file`: *Required.* Path to the file to upload, provided by an output of a task.

## Example Configuration

An example pipeline exists in the `example` directory.

### Resource

```
resource_types:
- name: azure-blobstore
  type: docker-image
  source:
    repository: pcfabr/azure-blobstore-resource

resources:
  - name: terraform-state
    type: azure-blobstore
    source:
      storage_account_name: {{storage_account_name}}
      storage_account_key: {{storage_account_key}}
      container: {{container}}
      versioned_file: terraform.tfstate
```

### Plan

```
- get: terraform-state
```

```
- put: terraform-state
  params:
    file: terraform-state/terraform.tfstate
```
