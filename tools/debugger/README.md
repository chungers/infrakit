# Debugger

The goal of this tools is to provide an easy way for us to gather information about their cluster so that we can help investigate what is going on.

## Phase 1: Collect and Report
The first phase is super simple, collect the data about the swarm, and have the user send it to us.

Items to collect
- number of nodes
- node information
    - Version
    - docker info
    - containers running
    - network
    - volumes
    - services

- Swarm info
    - VPC
    - subnets
    - IAM profile
    - Subnets
    - Availability Zones
    - Region
    - Security groups
    - Load balancers
        - open ports etc

User will run a simple command (curl | bash), this will download and start a container which has a debug script in it.
The script will gather the information, and save to a local file. the user then sends us that file.

## Phase 2: Prompt and Send
For phase 2, we expand on phase one, but prompt them a few questions, one of them is the ability to send the results automatically, so the user doesn't need to handle sending of the data.

1. Name
2. Email
3. Can we send the file automatically?

We will do this by collecting the answers via the shell, and then passing as parameters to the docker container.

Let the user review the data before we send, then automatically send it. If they don't want to us to automatically send it, they can still send it to us manually.

This will require the ability to have access to our systems.

How do we send? email, web service?

## Phase 3: Diagnose
With this phase, we start looking for common problems, and if we find them in the users setup, we give them suggestions on how to fix. We have a series of tests that are performed and if they fail, we can explain what is failing, and give tips on how to fix. Some tips might trigger more tests to help diagnose what is wrong.

We will not automatically fix any issues, the user will need to fix them on their own.

## Phase 4: Autocorrect
In this phase, we will give the ability to auto correct the issues that were found in Phase 3. We will prompt the user for confirmation before we do the fixes, and ideally have a way to back out the changes if required.

This is the most risky of all phases, and we don't want to move to this phase until we are sure we are confident with what we are doing.
