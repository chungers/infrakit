
import os
from boto import s3
from boto.exception import S3ResponseError
from datetime import datetime, timedelta

AWS_ACCESS_KEY_ID = os.environ.get('AWS_ACCESS_KEY_ID')
AWS_SECRET_ACCESS_KEY = os.environ.get('AWS_SECRET_ACCESS_KEY')
s3_bucket_name = "docker-for-aws"

REGIONS = ['us-west-1', 'us-west-2', 'us-east-1',
           'eu-west-1', 'eu-central-1', 'ap-southeast-1',
           'ap-northeast-1', 'ap-southeast-2', 'ap-northeast-2',
           'sa-east-1', 'ap-south-1', 'us-east-2', 'ca-central-1']

EXPIRE_AGE = 1
NOW = datetime.now()
EXPIRE_DATE = NOW - timedelta(EXPIRE_AGE)

delete_list = []

print("######## S3 bucket cleanup ########")
print("### NOW: {} ###".format(NOW))
print("### EXPIRE_DATE: {} ###".format(EXPIRE_DATE))

for region in REGIONS:
    print(u"Region: {} *****************************".format(region))
    conn = s3.connect_to_region(region)
    buckets = conn.get_all_buckets()
    for bucket in buckets:
        try:
            is_nightly = False
            delete = False
            has_files = False
            if (bucket.get_location() == region) or (bucket.get_location() == ''):
                print("  {}".format(bucket.name))
                tags = bucket.get_tags()[0]
                for tag in tags:
                    if tag.key == 'date':
                        print(u"    {}={}".format(tag.key, tag.value))
                        bucket_date = datetime.strptime(tag.value, "%m_%d_%Y")
                        if bucket_date < EXPIRE_DATE:
                            print("      * Expired")
                            delete = True
                            if len(bucket.get_all_keys(max_keys=1)) > 0:
                                has_files = True
                                print("      * There are Files ** Need to delete first. ")
                    elif tag.key == 'channel':
                        print(u"{}={}".format(tag.key, tag.value))
                        if tag.value == 'ddc-nightly':
                            print("      * Is nightly")
                            is_nightly = True
                if is_nightly and delete:
                    if has_files:
                        # TODO: This could take a long time and there code be data in there
                        # we want, so not deleting right now.
                        print("      # Delete keys in bucket first. Not Deleting....")
                    else:
                        print("      # Delete Bucket")
                        bucket.delete()
                        delete_list.append(bucket.name)
                        print("        !! Deleted bucket !!")
        except S3ResponseError:
            pass
        except Exception as exc:
            print(u"Error = {}".format(exc))
            pass

print("********")
print("The following buckets were deleted.")
for bkt in delete_list:
    print(" - {}".format(bkt))
print("********")
