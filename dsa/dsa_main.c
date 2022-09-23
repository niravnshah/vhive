#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdio.h>
#include "idxd_device.h"

// gcc -g -O0 dsa_main.c idxd_device.c -o dsa_c

int main(int argc, char** argv)
{
        int res;
        char *input, *output;

	dsa_setup("/dev/dsa/wq0.0");

	input = malloc(sizeof(char) * 4096);
	output = malloc(sizeof(char) * 4096);

	memset(input, 'A', 4096);
	memset(output, 0, 4096);

	dsa_memmove(input, output, 4096);

	res = memcmp(input, output, 4096);
	if (res == 0) {
		printf("dsa_memmove succeeded..!!\n");
	} else {
		printf("dsa_memmove failed..!!\n");
	}

	dsa_close();

        return 0;
}