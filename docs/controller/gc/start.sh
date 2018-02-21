{{/* =% sh %= */}}

# Since we are using file based leader detection, write the default name (manager1) to the leader file.
echo manager1 > ~/.infrakit/leader

# The simulators are started up with different names to mimic different resources
INFRAKIT_MANAGER_CONTROLLERS=gc \
infrakit plugin start manager:mystack vars group gc simulator:docker simulator:vm \
	 --log 5 --log-stack --log-debug-V 500 \
	 --log-debug-match-exclude \
	 --log-debug-match module=simulator/instance \
	 --log-debug-match module=rpc/internal \
	 --log-debug-match module=run/manager
