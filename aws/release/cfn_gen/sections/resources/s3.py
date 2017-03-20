from troposphere.s3 import Bucket


def add_s3_dtr_bucket(template):
    template.add_resource(Bucket(
        "DDCBucket",
        DeletionPolicy="Retain"
    ))
