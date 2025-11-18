package com.example.calculator;

import java.util.List;
import com.example.utils.MathHelper;

public class Calculator {
    public int add(int a, int b) {
        if (a > 0 && b > 0) {
            return a + b;
        }
        return 0;
    }

    public int subtract(int a, int b) {
        return a - b;
    }
}
