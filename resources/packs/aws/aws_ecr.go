package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	"github.com/cockroachdb/errors"
	"go.mondoo.com/cnquery/motor/providers/aws"
	"go.mondoo.com/cnquery/resources/library/jobpool"
	"go.mondoo.com/cnquery/resources/packs/core"
)

func (e *mqlAwsEcr) id() (string, error) {
	return "aws.ecr", nil
}

func (e *mqlAwsEcrRepository) id() (string, error) {
	return e.Arn()
}

func (e *mqlAwsEcrImage) id() (string, error) {
	id, err := e.RegistryId()
	if err != nil {
		return "", err
	}
	sha, err := e.Digest()
	if err != nil {
		return "", err
	}
	name, err := e.RepoName()
	if err != nil {
		return "", err
	}
	return id + "/" + name + "/" + sha, nil
}

func (e *mqlAwsEcr) GetImages() ([]interface{}, error) {
	obj, err := e.MotorRuntime.CreateResource("aws.ecr")
	if err != nil {
		return nil, err
	}
	ecr := obj.(AwsEcr)
	res := []interface{}{}

	repos, err := ecr.PublicRepositories()
	if err != nil {
		return nil, err
	}
	for i := range repos {
		images, err := repos[i].(AwsEcrRepository).Images()
		if err != nil {
			return nil, err
		}
		res = append(res, images...)
	}
	pRepos, err := ecr.PrivateRepositories()
	if err != nil {
		return nil, err
	}
	for i := range pRepos {
		images, err := pRepos[i].(AwsEcrRepository).Images()
		if err != nil {
			return nil, err
		}
		res = append(res, images...)
	}
	return res, nil
}

func (e *mqlAwsEcr) GetPrivateRepositories() ([]interface{}, error) {
	provider, err := awsProvider(e.MotorRuntime.Motor.Provider)
	if err != nil {
		return nil, err
	}
	res := []interface{}{}
	poolOfJobs := jobpool.CreatePool(e.getPrivateRepositories(provider), 5)
	poolOfJobs.Run()

	// check for errors
	if poolOfJobs.HasErrors() {
		return nil, poolOfJobs.GetErrors()
	}
	// get all the results
	for i := range poolOfJobs.Jobs {
		res = append(res, poolOfJobs.Jobs[i].Result.([]interface{})...)
	}

	return res, nil
}

func (e *mqlAwsEcr) getPrivateRepositories(provider *aws.Provider) []*jobpool.Job {
	tasks := make([]*jobpool.Job, 0)
	ctx := context.Background()

	regions, err := provider.GetRegions()
	if err != nil {
		return []*jobpool.Job{{Err: err}} // return the error
	}
	for ri := range regions {
		region := regions[ri]
		f := func() (jobpool.JobResult, error) {
			svc := provider.Ecr(region)
			res := []interface{}{}

			repoResp, err := svc.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{})
			if err != nil {
				return nil, err
			}
			for i := range repoResp.Repositories {
				r := repoResp.Repositories[i]
				mqlRepoResource, err := e.MotorRuntime.CreateResource("aws.ecr.repository",
					"arn", core.ToString(r.RepositoryArn),
					"name", core.ToString(r.RepositoryName),
					"uri", core.ToString(r.RepositoryUri),
					"registryId", core.ToString(r.RegistryId),
					"public", false,
					"region", region,
				)
				if err != nil {
					return nil, err
				}
				res = append(res, mqlRepoResource)
			}

			return jobpool.JobResult(res), nil
		}
		tasks = append(tasks, jobpool.NewJob(f))
	}
	return tasks
}

func (e *mqlAwsEcrRepository) GetImages() ([]interface{}, error) {
	name, err := e.Name()
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse name")
	}
	region, err := e.Region()
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse region")
	}
	public, err := e.Public()
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse public val")
	}
	at, err := awsProvider(e.MotorRuntime.Motor.Provider)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	mqlres := []interface{}{}
	if public {
		svc := at.EcrPublic(region)
		res, err := svc.DescribeImages(ctx, &ecrpublic.DescribeImagesInput{RepositoryName: &name})
		if err != nil {
			return nil, err
		}

		for i := range res.ImageDetails {
			image := res.ImageDetails[i]
			tags := []interface{}{}
			for i := range image.ImageTags {
				tags = append(tags, image.ImageTags[i])
			}
			mqlImage, err := e.MotorRuntime.CreateResource("aws.ecr.image",
				"digest", core.ToString(image.ImageDigest),
				"mediaType", core.ToString(image.ImageManifestMediaType),
				"tags", tags,
				"registryId", core.ToString(image.RegistryId),
				"repoName", name,
			)
			if err != nil {
				return nil, err
			}
			mqlres = append(mqlres, mqlImage)
		}
	} else {
		svc := at.Ecr(region)
		res, err := svc.DescribeImages(ctx, &ecr.DescribeImagesInput{RepositoryName: &name})
		if err != nil {
			return nil, err
		}
		for i := range res.ImageDetails {
			image := res.ImageDetails[i]
			tags := []interface{}{}
			for i := range image.ImageTags {
				tags = append(tags, image.ImageTags[i])
			}
			mqlImage, err := e.MotorRuntime.CreateResource("aws.ecr.image",
				"digest", core.ToString(image.ImageDigest),
				"mediaType", core.ToString(image.ImageManifestMediaType),
				"tags", tags,
				"registryId", core.ToString(image.RegistryId),
				"repoName", name,
			)
			if err != nil {
				return nil, err
			}
			mqlres = append(mqlres, mqlImage)
		}
	}
	return mqlres, nil
}

func (e *mqlAwsEcr) GetPublicRepositories() ([]interface{}, error) {
	provider, err := awsProvider(e.MotorRuntime.Motor.Provider)
	if err != nil {
		return nil, err
	}
	account, err := provider.Account()
	if err != nil {
		return nil, err
	}

	svc := provider.EcrPublic("us-east-1") // only supported for us-east-1
	res := []interface{}{}

	repoResp, err := svc.DescribeRepositories(context.TODO(), &ecrpublic.DescribeRepositoriesInput{RegistryId: &account.ID})
	if err != nil {
		return nil, err
	}
	for i := range repoResp.Repositories {
		r := repoResp.Repositories[i]
		mqlRepoResource, err := e.MotorRuntime.CreateResource("aws.ecr.repository",
			"arn", core.ToString(r.RepositoryArn),
			"name", core.ToString(r.RepositoryName),
			"uri", core.ToString(r.RepositoryUri),
			"registryId", core.ToString(r.RegistryId),
			"public", true,
			"region", "us-east-1",
		)
		if err != nil {
			return nil, err
		}
		res = append(res, mqlRepoResource)
	}

	return res, nil
}
