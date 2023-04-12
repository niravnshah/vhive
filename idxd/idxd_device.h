// SPDX-License-Identifier: GPL-2.0
#ifndef __IDXD_DEVICE_H__
#define __IDXD_DEVICE_H__

#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <errno.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/mman.h>

#define WQ_NUM_MAX 128
#define ARRAY_SIZE(x) (sizeof((x))/sizeof((x)[0]))
#define MAX_COMP_RETRY	2000000000
#define ERR printf
#if 0
#define ENTER printf("Entering %s\n", __func__)
#define EXIT printf("Exiting %s\n", __func__)
#else
#define ENTER
#define EXIT
#endif


void dsa_setup(char *path);
void dsa_close();
uint32_t dsa_wait_for_comp(struct dsa_hw_desc *hw_desc);
// uint32_t dsa_wait_for_comp_wrapper(void *hw_desc);
uint32_t dsa_memmove_sync(void *dst, void *src, uint32_t size);
uint32_t dsa_desc(struct dsa_hw_desc *hw_desc, uint sync);
uint32_t dsa_desc_wrapper(void *hw_desc, uint sync);

#endif
