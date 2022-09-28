// SPDX-License-Identifier: GPL-2.0
#ifndef __IDXD_DEVICE_H__
#define __IDXD_DEVICE_H__

#define WQ_NUM_MAX 128
#define ARRAY_SIZE(x) (sizeof((x))/sizeof((x)[0]))
#define MAX_COMP_RETRY	2000000000
#define ERR printf
#define ENTER printf("Entering %s\n", __func__)
#define EXIT printf("Exiting %s\n", __func__)

static void * wq;


// void dsa_setup(char *path);
// void dsa_close();
// int dsa_memmove(void *input, void *output, uint32_t size);

#endif
