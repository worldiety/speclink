// Command erp assembles the bounded contexts into a running application.
package main

import "example.com/erp/app/sales"

func main() {
	_ = sales.NewUseCases()
}
