using System;

namespace MyApp.Calculators
{
    public class Calculator
    {
        public int Add(int a, int b)
        {
            if (a > 0 && b > 0)
            {
                return a + b;
            }
            return 0;
        }

        public int Subtract(int a, int b)
        {
            return a - b;
        }
    }
}
