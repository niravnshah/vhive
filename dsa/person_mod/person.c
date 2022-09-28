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
#include "person_orig/person.h"

// gcc -o person person.c -Wl,-rpath=. libperson.so

struct APerson *get_person(const char *name, const char *long_name, struct APerson **outper){

    struct APerson *fmt = malloc(sizeof(struct APerson));
    fmt->name = name;
    fmt->long_name = long_name;

    struct APerson *temp = malloc(sizeof(struct APerson));
    temp->name = strdup("Static");
    temp->long_name = strdup("Cast");

    *outper = temp;

    return fmt;
};

struct APerson *get_person_only(const char *name, const char *long_name){

    printf("firstname = %s\n", name);
    printf("lastname = %s\n", long_name);
    struct APerson *fmt = malloc(sizeof(struct APerson));
    fmt->name = strdup(name);
    fmt->long_name = strdup(long_name);

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