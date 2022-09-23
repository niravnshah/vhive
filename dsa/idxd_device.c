// SPDX-License-Identifier: GPL-2.0
/* Copyright(c) 2019 Intel Corporation. All rights reserved. */
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/mman.h>

#include "dsa.h"
#include "idxd_device.h"

#define WQ_NUM_MAX 128

#define ENTER printf("Entering %s\n", __func__)
#define EXIT printf("Exiting %s\n", __func__)

void * wq;

int
dsa_memmove(void *input, void *output, uint32_t size)
{
	ENTER;
        struct dsa_hw_desc desc = {};
        struct dsa_completion_record comp = {};
	uint32_t retry = 0;

        desc.src_addr = (uint64_t)input;
        desc.dst_addr = (uint64_t)output;
        desc.xfer_size = 4096;
        desc.opcode = DSA_OPCODE_MEMMOVE;
        desc.completion_addr = (uint64_t)&comp;
        desc.flags = IDXD_OP_FLAG_CRAV | IDXD_OP_FLAG_RCR;

        dsa_desc_submit(wq, 0, (void*)&desc);

        while (comp.status == 0 && retry++ < MAX_COMP_RETRY) {
                umonitor(&comp);
		// if (comp.status == 0) {
		// 	umwait();
		// }
        }

	if (!memcmp(input, output, size))
		printf("%s : dsa_memmove succeeeded!\n", __func__);
	else
		printf("%s : dsa_memmove failed!\n", __func__);

	EXIT;
	return 0;
}

void
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

void
dsa_close()
{
	ENTER;
	munmap(wq, 0x1000);
	EXIT;
}
