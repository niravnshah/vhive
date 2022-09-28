/*
 * person.c
 * Copyright (C) 2019 Tim Hughes
 *
 * Distributed under terms of the MIT license.
 */

#include <stdlib.h>
#include <stdio.h>
#include "person.h"


struct APerson *get_person(const char *name, const char *long_name){

    struct APerson *fmt = malloc(sizeof(struct APerson));
    fmt->name = name;
    fmt->long_name = long_name;

    printf("Here ia m...\n");

    return fmt;
};