# Running Acceptance Tests
Running the acceptance tests requires running against an Azure storage account.
Once you have a storage account that can be used by the acceptance tests you
need to supply two environment variables:

```
export TEST_STORAGE_ACCOUNT_NAME=<your-storage-account-name>
export TEST_STORAGE_ACCOUNT_KEY=<your-storage-account-key>
```

Now you can run the tests using ginkgo:

```
# Assuming you are in the repo's root directory
ginkgo acceptance
```
