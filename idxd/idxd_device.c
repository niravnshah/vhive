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

void * wq;

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

uint32_t
dsa_memmove_sync(void *dst, void *src, uint32_t size)
{
	ENTER;
        struct dsa_hw_desc desc __attribute__ ((aligned (64))) = {};
        struct dsa_completion_record comp __attribute__ ((aligned (32))) = {};

        desc.src_addr = (uint64_t)src;
        desc.dst_addr = (uint64_t)dst;
        desc.xfer_size = size;
        desc.opcode = DSA_OPCODE_MEMMOVE;
        desc.completion_addr = (uint64_t)&comp;
        desc.flags = IDXD_OP_FLAG_CRAV | IDXD_OP_FLAG_RCR;

        dsa_memmove_desc(&desc, 1);

	EXIT;
	return comp.status != 1;
}

uint32_t
dsa_memmove_desc(struct dsa_hw_desc *hw_desc, uint sync)
{
	ENTER;

        dsa_desc_submit(wq, 0, hw_desc);

	if (sync) {
	        struct dsa_completion_record *comp =
			(struct dsa_completion_record *)hw_desc->completion_addr;
        	dsa_wait_for_comp(hw_desc);
		EXIT;
		return comp->status != 1;
	}
	EXIT;
	return 0;
}

uint32_t
dsa_memmove_desc_wrapper(void *hw_desc, uint sync)
{
	return dsa_memmove_desc((struct dsa_hw_desc *)(hw_desc), sync);
}

uint32_t
dsa_wait_for_comp(struct dsa_hw_desc *hw_desc)
{
	ENTER;
	uint32_t retry = 0;
        struct dsa_completion_record *comp =
		(struct dsa_completion_record *)hw_desc->completion_addr;

        while (comp->status == 0 && retry++ < MAX_COMP_RETRY) {
                umonitor(&comp);
		// if (comp.status == 0) {
		// 	umwait();
		// }
        }

	if (retry >= MAX_COMP_RETRY) {
		ERR("Desc timeout!!\n");
		EXIT;
		return 1;
	}
	EXIT;
	return 0;
}

// uint32_t
// dsa_wait_for_comp_wrapper(void *hw_desc)
// {
// 	return dsa_wait_for_comp((struct dsa_hw_desc *)(hw_desc));
// }

