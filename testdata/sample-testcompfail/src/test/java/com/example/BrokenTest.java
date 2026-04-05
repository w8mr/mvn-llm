package com.example;

import org.junit.Test;

public class BrokenTest {
    @Test
    public void testSomething() {
        // Missing semicolon will cause compile error
        int x = 1
        System.out.println(x);
    }
}