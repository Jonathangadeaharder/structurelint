#include <iostream>
#include "calculator.h"

int Calculator::add(int a, int b) {
    if (a > 0 && b > 0) {
        return a + b;
    }
    return 0;
}

int Calculator::subtract(int a, int b) {
    return a - b;
}
