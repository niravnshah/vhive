#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <linux/idxd.h>

#include "greet.h"
#include "person.h"
#include "idxd_device.h"
#include "dsa.h"

//  gcc -o greet greet.c
// gcc -g -O0 -o greet greet.c -Wl,-rpath=. libperson.so
// gcc -g -O0 -o greet idxd_device.c greet.c -Wl,-rpath=. libperson.so

int greet(const char *name, int year, char *out)
{
        int n;

        n = sprintf(out, "Buhahahahaha %s from year %d!, We come in peace :)", name, year);

        APerson * of;
        APerson * newPer;
        of = get_person("tim", "tim hughes",&newPer);
        printf("Hello C world: My name is %s, %s.\n", of->name, of->long_name);
        printf("New Per: My name is %s, %s.\n", newPer->name, newPer->long_name);

        return n;
}

void start_greet()
{
        char *name = "John";
        int y = 2022;
        char out[1024] = {0};
        greet(name, y, out);
        printf("%s", out);
}

// int main(int argc, char** argv)
// {
//         start_greet();
// }