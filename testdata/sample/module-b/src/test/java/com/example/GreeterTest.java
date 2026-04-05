package com.example;

import org.junit.Test;
import static org.junit.Assert.*;

public class GreeterTest {
    @Test
    public void testGreet() {
        Greeter greeter = new Greeter();
        assertEquals("Hello, World!", greeter.greet("World"));
    }
}