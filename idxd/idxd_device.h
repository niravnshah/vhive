// SPDX-License-Identifier: GPL-2.0
#ifndef __IDXD_DEVICE_H__
#define __IDXD_DEVICE_H__

#include <stdint.h>

#define ARRAY_SIZE(x) (sizeof((x))/sizeof((x)[0]))
#define ERR printf
#define MAX_COMP_RETRY	2000000000

void dsa_setup(char *path);
void dsa_close();
int dsa_memmove(void *input, void *output, uint32_t size);

#endif
