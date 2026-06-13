import{a as h,t as u}from"../chunks/BpVn7OHi.js";import"../chunks/DwZJtKG4.js";import{p as j,g as L,$ as M,s as e,c as r,r as o,t as Y,i as $}from"../chunks/CXOfwQjV.js";import{s as B}from"../chunks/DljT5etN.js";import{e as C,i as D}from"../chunks/B-jJNYo9.js";import{h as U}from"../chunks/D9A97xFR.js";import{s as A}from"../chunks/CYhT1aF9.js";import{d as F}from"../chunks/BAO5_qdH.js";import{i as E}from"../chunks/C_gcfv8f.js";import{b as q}from"../chunks/BsB74Sq8.js";var N=u(`<div style="
						display: flex;
						align-items: center;
						gap: var(--spacing-md);
						padding: var(--spacing-sm) var(--spacing-md);
						background-color: var(--color-surface-card);
						border: 1px solid var(--color-border-subtle);
						border-radius: 6px;
					"><span style="
							width: 24px;
							height: 24px;
							border-radius: 999px;
							background-color: var(--color-accent-brand);
							color: #ffffff;
							font-size: var(--font-size-label);
							font-weight: 600;
							display: flex;
							align-items: center;
							justify-content: center;
							flex-shrink: 0;
						"></span> <span style="font-size: var(--font-size-body); color: var(--color-text-primary);"> </span></div>`),V=u(`<div style="
		flex: 1;
		padding: var(--spacing-3xl) var(--spacing-lg);
		max-width: 800px;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		gap: var(--spacing-2xl);
	"><div style="display: flex; flex-direction: column; gap: var(--spacing-sm);"><h1 style="
				font-size: var(--font-size-display);
				font-weight: 600;
				color: var(--color-text-primary);
				margin: 0;
			">Your speech is ready.</h1> <p style="font-size: var(--font-size-body); color: var(--color-text-secondary); margin: 0;">No conflicts remain. Run the command below to build your installer, or download the resolved
			speech YAML to build later.</p></div> <div style="display: flex; flex-direction: column; gap: var(--spacing-sm);"><h2 style="
				font-size: var(--font-size-heading);
				font-weight: 600;
				color: var(--color-text-primary);
				margin: 0;
			">Build command</h2> <pre style="
				font-family: var(--font-mono);
				font-size: var(--font-size-code);
				line-height: var(--line-height-code);
				background-color: var(--color-surface-card);
				border: 1px solid var(--color-border-subtle);
				border-radius: 6px;
				padding: var(--spacing-md);
				overflow-x: auto;
				margin: 0;
				color: var(--color-text-primary);
			"></pre></div> <div style="display: flex; flex-direction: column; gap: var(--spacing-sm);"><h2 style="
				font-size: var(--font-size-heading);
				font-weight: 600;
				color: var(--color-text-primary);
				margin: 0;
			">Build stages</h2> <div style="display: flex; flex-direction: column; gap: var(--spacing-xs);"></div></div> <div style="display: flex; flex-direction: column; gap: var(--spacing-sm);"><h2 style="
				font-size: var(--font-size-heading);
				font-weight: 600;
				color: var(--color-text-primary);
				margin: 0;
			">Resolved speech</h2> <pre style="
				font-family: var(--font-mono);
				font-size: var(--font-size-code);
				line-height: var(--line-height-code);
				background-color: var(--color-surface-card);
				border: 1px solid var(--color-border-subtle);
				border-radius: 6px;
				padding: var(--spacing-md);
				overflow-x: auto;
				margin: 0;
				color: var(--color-text-primary);
				max-height: 320px;
				overflow-y: auto;
			"></pre></div> <div style="display: flex; gap: var(--spacing-md); flex-wrap: wrap;"><button style="
				display: inline-flex;
				align-items: center;
				justify-content: center;
				min-height: var(--min-height-touch);
				padding: 0 var(--spacing-xl);
				background-color: var(--color-accent-brand);
				color: #ffffff;
				font-size: var(--font-size-body);
				font-weight: 600;
				border: none;
				border-radius: 6px;
				cursor: pointer;
			">Download Resolved Speech</button> <a style="
				display: inline-flex;
				align-items: center;
				justify-content: center;
				min-height: var(--min-height-touch);
				padding: 0 var(--spacing-xl);
				border: 1px solid var(--color-border-subtle);
				border-radius: 6px;
				color: var(--color-text-secondary);
				font-size: var(--font-size-body);
				text-decoration: none;
			">Back to Debate</a></div></div>`);function Z(b,y){j(y,!1);const z=[{key:"resolve",label:"Settling the Debate"},{key:"translate",label:"Finding Your Foundation's Voice"},{key:"build",label:"Writing the Final Argument"},{key:"package",label:"Sealing the Speech"}],w="debateos build --speech resolved-speech.yaml",p=`schema: 1
foundation: arch
applied:
  - OM-001
  - OM-006
  - OM-097
  - OM-099
dropped: []
explanations:
  - rule: no-conflict
    text: "All required opinions are compatible."
`;function k(){const t=new Blob([p],{type:"text/yaml"}),i=URL.createObjectURL(t),a=document.createElement("a");a.href=i,a.download="resolved-speech.yaml",a.click(),URL.revokeObjectURL(i)}E();var n=V();U(t=>{M.title="Export — DebateOS"});var l=e(r(n),2),_=e(r(l),2);_.textContent=w,o(l);var s=e(l,2),v=e(r(s),2);C(v,5,()=>z,D,(t,i,a)=>{var c=N(),g=r(c);g.textContent=a+1;var x=e(g,2),S=r(x,!0);o(x),o(c),Y(()=>B(S,$(i).label)),h(t,c)}),o(v),o(s);var d=e(s,2),O=e(r(d),2);O.textContent=p,o(d);var f=e(d,2),m=r(f);m.__click=k;var R=e(m,2);A(R,"href",`${q??""}/debate/`),o(f),o(n),h(b,n),L()}F(["click"]);export{Z as component};
