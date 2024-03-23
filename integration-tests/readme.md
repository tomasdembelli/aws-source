This document investigates how to improve the test quality across the project.

Current issues:
- The project has no integration tests in a sense of validating the linked item queries.
- The unit tests are using hand-crafted data, which might not reflect the real data.

# Thoughts on the existing unit tests:
- Even though the unit tests are using hand-crafted data, they are still valuable for validating the linked item query structure.
- In unit tests, it is possible to set a source health status, which is not possible in integration tests.
- They are cheap to run and fast to develop.

So, I think we should keep them.
A better test improvement can be achieved by adding integration tests.

# What should we test in integration tests?
- Get, List and Search methods of each source.
- Run each linked item query and validate the result.
- Basically, only happy paths scenarios should be tested.

# How an ideal integration test should look like?
- They should be grouped by source family, i.e., ec2, s3, etc. Otherwise, maintaining the test suite can be very challenging and definitely not ideal.
- But, they should also include linked items even if they are from different source family, i.e., including lambda function for a sqs test scenario.
- Each test should be independent of each other. Otherwise, the test suite can be very fragile.
- Each test should start with creation of the relevant sources. Sources should be linked to each other for the linked item queries. (Adding sqs que to a lambda function, etc.)

Creating sources is not a trivial task. But it will help author to better understand the source and improve the overall code quality.
Test sources should be created with AWS Go SDK. This will save contributors, who are not familiar with Terraform, from the burden of learning Terraform or other alternatives.

# Is localstack a good choice for integration tests?
Localstack is not a good choice for integration tests.
It lacks some of the sources that we have in the project. For example, it does not support the `DescribeBackup` endpoint for DynamoDB. See [here](https://docs.localstack.cloud/references/coverage/coverage_dynamodb/) for details.
Also, it has community and Pro versions. To be able to get some benefits, it is inevitable to use the Pro version. See [here](https://localstack.cloud/pricing/) for pricing.

So, we should use the real AWS services for integration tests.

# When and how to run integration tests?
We can run them after a PR approval, not after every commit in a PR.
We can run only the impacted tests. So, if there is a change in a ec2 source, we can run only the ec2 tests.

This will still keep the development process light and fast, but also keep the cost of the tests at a minimum.

Realistically we should run them couple of times a day, but it comes down the cost of the tests.

# Next steps
If we can generally agree on above points, then we can start developing the integration test suit with the localstack community edition. 
Localstack supports almost all endpoints for s3, sqs and lambda. We can start with them.

After agreeing on the test framework then we can start using the real AWS services.