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

static int
dsa_memmove(void *dst, void *src, uint32_t size)
{
	ENTER;
        struct dsa_hw_desc desc __attribute__ ((aligned (64))) = {};
        struct dsa_completion_record comp __attribute__ ((aligned (32))) = {};
	uint32_t retry = 0;

        desc.src_addr = (uint64_t)dst;
        desc.dst_addr = (uint64_t)src;
        desc.xfer_size = size;
        desc.opcode = DSA_OPCODE_MEMMOVE;
        desc.completion_addr = (uint64_t)&comp;
        desc.flags = IDXD_OP_FLAG_CRAV | IDXD_OP_FLAG_RCR;

        /*dsa_desc_submit(wq, 0, (void*)&desc);

        while (comp.status == 0 && retry++ < MAX_COMP_RETRY) {
                umonitor(&comp);
		// if (comp.status == 0) {
		// 	umwait();
		// }
        }*/

	EXIT;
	return 0;
}

