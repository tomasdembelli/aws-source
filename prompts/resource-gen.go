package prompts

/*

I want to write a new source for the aws apigateway domain-name.
- Define the package, file and source name. Reference to README.md .
- Choose the best backend source type, use the files provided next to source names: AlwaysGetSource: always_get_source.go , DescribeOnlySource: describe_source.go , GetListSource: get_list_source.go . Also investigate the available methods for the source on the remote aws-sdk-v2 repository: GitHub - aws/aws-sdk-go-v2: AWS SDK for the Go programming language. Make sure that you understand well the docs written for the backend types.
- It is very important to find the  `UniqueAttribute` when creating the `sdp.Item` for the source. Sometimes it can be very straigtforward like an `id`, i.e., key.go . Also, it can be a composite value with a custom key as in the grant.go .
- Find the tags for a resource. In most cases the returned object from AWS has tags on it, i.e., connection.go . Sometimes we need to make a separate call to get the tags: key.go . If we have to make a call verify tha the `aws.Item` has all the necessary information to make this call. If it doesn't have, simply do not try to get the tags.
- For the linked item queries: Check all the first level and nested attributes of the AWS object. If they contain a source identifier like an ID or similar for a composite ID, create a linked item for them. Sometimes we can create a linked item for the nonexisting sources in the overmind aws-source repository for future readiness. Also, we should check whether we can use the unique attribute of the source that we are currently creating for linking to other sources (existing or nonexisting) i.e.:  key.go .
- The blast propogation for the linked items. If any change in the linked item has an impact on the source that we are creating, then IN should be TRUE. If any change on the source we are creating can have an impact on the linked item then OUT should be TRUE. If in doubt go for TRUE to be on the safe side. 
- Do not forget to add the linked item doc comments.
- If the AWS response has any state related attribute, use that to add a Health state to sdp.Item, example: key.go .
- Do not forget to add doc comments to the source itself which starts with `go:generate docgen ../../docs-data.
- Always verify that the parameter you pass for a method call exists on the type that we are populating for an SDK call. 
*/
