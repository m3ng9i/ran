#!/usr/bin/env python3

import os, time, subprocess, getopt, sys


def runCmd(cmd):
    p = subprocess.Popen(cmd, shell = True, stdout = subprocess.PIPE, stderr = subprocess.PIPE)
    stdout = p.communicate()[0].decode('utf-8').strip()
    return stdout


# Get last tag.
def lastTag():
    return runCmd('git describe --abbrev=0 --tags')


# Get current branch name.
def branch():
    return runCmd('git rev-parse --abbrev-ref HEAD')


# Get last git commit id.
def lastCommitId():
    return runCmd('git log --pretty=format:"%h" -1')


# Get package name in the current directory.
# E.g. github.com/m3ng9i/ran
def packageName():
    return runCmd("go list")


pkgName = ""


# Assemble build command.
def buildCmd():
    buildFlag = []

    version = lastTag()
    if version != "":
        buildFlag.append("-X '{}/global._version_={}'".format(pkgName, version))

    branchName = branch()
    if branchName != "":
        buildFlag.append("-X '{}/global._branch_={}'".format(pkgName, branchName))

    commitId = lastCommitId()
    if commitId != "":
        buildFlag.append("-X '{}/global._commitId_={}'".format(pkgName, commitId))

    # current time
    buildFlag.append("-X '{}/global._buildTime_={}'".format(pkgName, time.strftime("%Y-%m-%d %H:%M %z")))

    return 'go build -ldflags "{}"'.format(" ".join(buildFlag))


validOSArch = {
    "darwin":       ["386", "amd64", "arm", "arm64"],
    "dragonfly":    ["amd64"],
    "freebsd":      ["386", "amd64", "arm"],
    "linux":        ["386", "amd64", "arm", "arm64", "ppc64", "ppc64le"],
    "netbsd":       ["386", "amd64", "arm"],
    "openbsd":      ["386", "amd64", "arm"],
    "plan9":        ["386", "amd64"],
    "solaris":      ["amd64"],
    "windows":      ["386", "amd64"],
}


# Check if GOOS and GOARCH is valid combinations.
# Learn more at https://golang.org/doc/install/source
def isValidOSArch(goos, goarch):
    os = validOSArch.get(goos)
    if os is None:
        return False

    if goarch in os:
        return True

    return False


# Build binary for current OS and architecture
def build():
    if subprocess.call(buildCmd(), shell = True) == 0:
        print("Build finished.")


# Build binaries for specify OS and architecture
# pairs: valid GOOS/GOARCH pairs
# filePrefix: filename prefix used in output binaries
def buildPlatform(pairs, filePrefix):
    cmd = buildCmd()

    for p in pairs:
        filename = "{}_{}_{}".format(filePrefix, p[0], p[1])
        if p[0] == "windows":
            filename += ".exe"

        c = "GOOS={} GOARCH={} {} -o {}".format(p[0], p[1], cmd, filename)
        if subprocess.call(c, shell = True) == 0:
            print("Build finished: {}".format(filename))
        else:
            # build error
            return

    print("All build finished.")


usage = """Go binary builder

Usage:

    ./build.py [GOOS/GOARCH pairs...]
    ./build.py [-h, --help]

Examples:

    1. Build binary for current OS and architecture:

        ./build.py

    2. Build binary for windows/386:

        ./build.py windows/386

    3. Build binaries for windows/386 and linux/386:

        ./build.py windows/386 linux/386

    4. Build binaries for linux/386 and linux/amd64:

        ./build.py linux/386 linux/amd64"""


errmsg = "Arguments are not valid GOOS/GOARCH pairs, use -h for help"


def main():

    global pkgName
    pkgName = packageName()
    if pkgName == "":
        sys.exit("Can not get package name, you must run this command under a go import path.")

    if len(sys.argv) <= 1:
        build()
        return

    validPairs = []

    for arg in sys.argv[1:]:
        arg = arg.lower()

        if arg in ["-h", "--help"]:
            print(usage)
            return

        pairs = arg.split("/")
        if len(pairs) != 2:
            sys.exit(errmsg)

        if isValidOSArch(pairs[0], pairs[1]) is False:
            sys.exit(errmsg)

        validPairs.append(pairs)

    buildPlatform(validPairs, "ran")


if __name__ == "__main__":
    main()
