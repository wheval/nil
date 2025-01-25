import CookieConsent, {} from "react-cookie-consent";
import { CustomRuntimeProvider } from "../components/CustomChatRuntimeProvider";
import "@assistant-ui/react/styles/index.css";
import "@assistant-ui/react/styles/modal.css";
import { GoogleReCaptchaProvider } from "react-google-recaptcha-v3";

// Default implementation, that you can customize
export default function Root({ children }) {
  const buttonStyle = {
    display: "flex",
    flexDirection: "row",
    alignItems: "center",
    backgroundColor: "#f2f2f2",
    padding: "0.75rem",
    borderRadius: "5px",
    fontFamily: "HelveticaNeue",
  };
  const contentStyle = {
    fontFamily: "HelveticaNeue",
    color: "#f2f2f2",
  };

  const style = {
    backgroundColor: "#181A1B",
  };
  return (
    <GoogleReCaptchaProvider reCaptchaKey="6Lf1w7kqAAAAAD-adI2XAasjjGJz7QC_DEBYc9Ft">
      <CustomRuntimeProvider>
        <CookieConsent
          location="bottom"
          buttonText="I understand"
          buttonStyle={buttonStyle}
          contentStyle={contentStyle}
          overlay={false}
          style={style}
        >
          This website tracks page views and other usage statistics. No personal information is
          collected.
        </CookieConsent>
        {children}
      </CustomRuntimeProvider>
    </GoogleReCaptchaProvider>
  );
}
