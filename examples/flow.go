
package main

import (
	"fmt"
	"github.com/sjqzhang/fsm"
)

func main() {


	const FlowCampus=`
AddNewEmployeeToSystemWithoutEmail: init → EmloyeePendingInSystem
NotifyManagerToAddReportingManagerAndTechLeaderForNewEmployee: EmloyeePendingInSystem → WaitForManagerToAddReportingManagerAndTechLeaderForNewEmployee
ManagerAddReportingManagerAndTechLeaderForNewEmployee: WaitForManagerToAddReportingManagerAndTechLeaderForNewEmployee → FinishAddReportingManagerAndTechLeaderToNewEmployee
EmloyeeHiredAtAgreedTime: FinishAddReportingManagerAndTechLeaderToNewEmployee → WaitForHRInsertEmailForNewEmployee
HRInsertEmailForNewEmployee: WaitForHRInsertEmailForNewEmployee → EmployeeActiveInSystem
`

	m,e:=fsm.NewFSMFromTemplate("init",FlowCampus,fsm.Callbacks{})
	if e!=nil {
		fmt.Println(e)
		return
	}
	//fmt.Println(m.AvailableTransitions())


	fmt.Println(fsm.Visualize(m))

}