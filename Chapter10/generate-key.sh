#!/bin/bash

ssh-keygen -t rsa -b 2048 -f $(pwd)/ssh.key -N ""

ssh-add ssh.key
