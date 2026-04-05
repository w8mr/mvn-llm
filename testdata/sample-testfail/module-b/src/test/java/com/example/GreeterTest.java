package com.example;

import org.junit.Test;
import static org.junit.Assert.*;

public class GreeterTest {
    @Test
    public void testGreet() {
        Greeter g = new Greeter();
        assertEquals("Hello, World!", g.greet("World"));
    }
}