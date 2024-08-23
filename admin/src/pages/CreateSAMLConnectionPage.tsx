import React, { ReactNode, useState } from "react";
// import { Steps } from "@/components/Steps";
import { Link } from "react-router-dom";
import { CheckIcon, ChevronRightIcon } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";

interface IDP {
  id: string;
  displayName: string;
  subtitle: string;
  steps: Step[];
}

const IDPS: IDP[] = [
  {
    id: "okta",
    displayName: "Okta",
    subtitle: "Set up a SAML connection with your corporate Okta.",
    steps: [
      {
        displayName: "Create application",
      },
      {
        displayName: "Configure application",
      },
      {
        displayName: "Finish creating application",
      },
      {
        displayName: "Copy Metadata URL",
      },
      {
        displayName: "Assign users",
      },
    ],
  },
  {
    id: "google",
    displayName: "Google",
    subtitle: "Set up a SAML connection with your Google Workspace.",
    steps: [],
  },
  {
    id: "entra",
    displayName: "Microsoft Entra",
    subtitle:
      "Set up a SAML connection with your Microsoft Entra (formerly known as Azure Active Directory).",
    steps: [],
  },
  {
    id: "other",
    displayName: "Other Identity Provider",
    subtitle: "Set up a SAML connection with any other identity provider.",
    steps: [],
  },
];

export function CreateSAMLConnectionPage() {
  const [idpId, setIdpId] = useState<string | undefined>(undefined);
  const idp = IDPS.find(({ id }) => id === idpId);
  const [currentStep, setCurrentStep] = useState(0);

  return (
    <>
      <nav className="border-b border-gray-200 bg-white">
        {idp ? (
          <Steps steps={idp.steps} currentStep={currentStep} />
        ) : (
          <div className="mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex h-16 justify-between">
              <div className="flex text-sm items-center"></div>
              <div className="flex text-sm items-center">
                SAML Connection Setup
              </div>
              <div className="flex text-sm items-center">
                <Link to="/">Cancel</Link>
              </div>
            </div>
          </div>
        )}
      </nav>

      <NarrowContainer>
        {idp ? (
          <div className="mt-8">
            {idpId === "okta" && currentStep === 0 && (
              <OktaCreateAppStep onNextStep={() => setCurrentStep(1)} />
            )}
            {idpId === "okta" && currentStep === 1 && (
              <OktaConfigureAppStep onNextStep={() => setCurrentStep(1)} />
            )}
          </div>
        ) : (
          <>
            <h1 className="mt-8 text-xl font-semibold">
              What Identity Provider do you use?
            </h1>

            <div className="mt-4 overflow-hidden bg-white shadow sm:rounded-md">
              <ul role="list" className="divide-y divide-gray-200">
                {IDPS.map((idp) => (
                  <li
                    key={idp.id}
                    onClick={() => setIdpId(idp.id)}
                    className="px-4 py-4 sm:px-6 text-sm flex items-center justify-between cursor-pointer"
                  >
                    <div>
                      <div className="font-medium text-gray-900">
                        {idp.displayName}
                      </div>
                      <div className="text-gray-500">{idp.subtitle}</div>
                    </div>

                    <div>
                      <ChevronRightIcon className="text-gray-700" />
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          </>
        )}
      </NarrowContainer>
    </>
  );
}

function NarrowContainer({ children }: { children?: ReactNode }) {
  return (
    <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-3xl">{children}</div>
    </div>
  );
}

interface Step {
  displayName: string;
}

function Steps({ steps, currentStep }: { steps: Step[]; currentStep: number }) {
  return (
    <div className="flex w-full justify-center">
      <div className="flex items-center">
        {steps.map((step, index) => (
          <div key={index}>
            {index < currentStep && (
              <div className="group flex w-full items-center">
                <span className="flex items-center px-6 py-4 text-sm font-medium">
                  <span className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-indigo-600 group-hover:bg-indigo-800">
                    <CheckIcon className="h-6 w-6 text-white" />
                  </span>
                  <span className="ml-4 text-sm font-medium text-gray-900">
                    {step.displayName}
                  </span>
                </span>
              </div>
            )}

            {index === currentStep && (
              <div
                aria-current="step"
                className="flex items-center px-6 py-4 text-sm font-medium"
              >
                <span className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full border-2 border-indigo-600">
                  <span className="text-indigo-600">{index + 1}</span>
                </span>
                <span className="ml-4 text-sm font-medium text-indigo-600">
                  {step.displayName}
                </span>
              </div>
            )}

            {index > currentStep && (
              <div className="group flex items-center">
                <span className="flex items-center px-6 py-4 text-sm font-medium">
                  <span className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full border-2 border-gray-300 group-hover:border-gray-400">
                    <span className="text-gray-500 group-hover:text-gray-900">
                      {index + 1}
                    </span>
                  </span>
                  <span className="ml-4 text-sm font-medium text-gray-500 group-hover:text-gray-900">
                    {step.displayName}
                  </span>
                </span>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

function OktaCreateAppStep({ onNextStep }: { onNextStep: () => void }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Create an Okta application</CardTitle>
        <CardDescription>
          Create a new Okta application that will let you log into our
          application.
        </CardDescription>
      </CardHeader>

      <CardContent>
        <img src="/okta_create_app.gif" />

        <div className="text-sm mt-4">
          <p>Create a new Okta application:</p>

          <ol className="mt-2 list-decimal list-inside space-y-1">
            <li>Go to Applications &gt; Applications in the sidebar.</li>
            <li>Click "Create App Integration"</li>
            <li>Choose "SAML 2.0"</li>
            <li>Click "Next"</li>
          </ol>
        </div>

        <div className="flex justify-end">
          <Button onClick={onNextStep}>Next: Configure application</Button>
        </div>
      </CardContent>
    </Card>
  );
}

function OktaConfigureAppStep({ onNextStep }: { onNextStep: () => void }) {
  const [subStep, setSubStep] = useState(0);

  return (
    <Card>
      {subStep === 0 && (
        <>
          <CardHeader>
            <CardTitle>Configure app name</CardTitle>
            <CardDescription>
              Configure the name of your Okta app. This is shown to employees.
            </CardDescription>
          </CardHeader>

          <CardContent>
            <img src="/okta_configure_app_name.gif" />

            <div className="text-sm mt-4">
              <p>Give the new Okta application a name.</p>
            </div>

            <div className="flex justify-end">
              <Button onClick={() => setSubStep(1)}>
                Next: Configure SAML Single Sign-on URL
              </Button>
            </div>
          </CardContent>
        </>
      )}
    </Card>
  );
}
