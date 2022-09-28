// SPDX-License-Identifier: GPL-2.0
/* Copyright(c) 2019 Intel Corporation. All rights reserved. */
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <errno.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/mman.h>

#include "dsa.h"
#include "idxd_device.h"

static void
dsa_setup(char *path)
{
	ENTER;
	int fd;

	fd = open(path, O_RDWR);
	if (fd < 0) {
		ERR("File open error %s: %s\n", path, strerror(errno));
		return;
	}

	wq = mmap(NULL, 0x1000, PROT_WRITE, MAP_SHARED | MAP_POPULATE, fd, 0);
	if (wq == MAP_FAILED) {
		ERR("mmap error: %s", strerror(errno));
		close(fd);
		return;
	}

	close(fd);
	EXIT;
}

static void
dsa_close()
{
	ENTER;
	munmap(wq, 0x1000);
	EXIT;
}
