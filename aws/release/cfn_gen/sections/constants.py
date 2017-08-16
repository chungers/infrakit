DDC_INSTANCE_TYPES = [
    "m4.large", "m4.xlarge", "m4.2xlarge", "m4.4xlarge", "m4.10xlarge",
    "m3.medium", "m3.large", "m3.xlarge", "m3.2xlarge", "c4.large",
    "c4.xlarge", "c4.2xlarge", "c4.4xlarge", "c4.8xlarge", "c3.large",
    "c3.xlarge", "c3.2xlarge", "c3.4xlarge", "c3.8xlarge", "r3.large",
    "r3.xlarge", "r3.2xlarge", "r3.4xlarge", "r3.8xlarge", "r4.large",
    "r4.xlarge", "r4.2xlarge", "r4.4xlarge", "r4.8xlarge", "r4.16xlarge",
    "i2.xlarge", "i2.2xlarge", "i2.4xlarge", "i2.8xlarge", "i3.large",
    "i3.xlarge", "i3.2xlarge", "i3.4xlarge", "i3.8xlarge", "i3.16xlarge",
]

DDC_WORKER_INSTANCE_TYPES = [
    "t2.small", "t2.medium", "t2.large", "t2.xlarge", "t2.2xlarge",
] + DDC_INSTANCE_TYPES

ALL_INSTANCE_TYPES = [
    "t2.micro", "t2.small", "t2.medium", "t2.large", "t2.xlarge", "t2.2xlarge",
] + DDC_INSTANCE_TYPES
