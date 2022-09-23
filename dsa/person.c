/*
 * person.c
 * Copyright (C) 2019 Tim Hughes
 *
 * Distributed under terms of the MIT license.
 */

// gcc -o libperson.so -Wall -g -shared -fPIC person.c -O0

#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include "person.h"

// gcc -o person person.c -Wl,-rpath=. libperson.so

APerson *get_person(const char *name, const char *long_name, APerson **outper){

    APerson *fmt = malloc(sizeof(APerson));
    fmt->name = name;
    fmt->long_name = long_name;

    APerson *temp = malloc(sizeof(APerson));
    temp->name = strdup("Static");
    temp->long_name = strdup("Cast");

    *outper = temp;

    return fmt;
};

// int main(int argc, char** argv)
// {
//     APerson * of;
//     APerson * newPerson;
//     of = get_person("tim", "tim hughes", &newPerson);
//     printf("Hello C world: My name is %s, %s.\n", of->name, of->long_name);
//     printf("New Person: My name is %s, %s.\n", newPerson->name, newPerson->long_name);
//     return 0;
// }